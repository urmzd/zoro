package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/urmzd/zoro/api/internal/model"
)

type KnowledgeStore interface {
	AddEpisode(ctx context.Context, req model.EpisodeRequest) (*model.EpisodeResponse, error)
	SearchFacts(ctx context.Context, query string, groupID string) (*model.SearchFactsResponse, error)
	GetGraph(ctx context.Context, limit int) (*model.GraphData, error)
	GetNode(ctx context.Context, id string, depth int) (*model.NodeDetail, error)
	EnsureSchema(ctx context.Context) error
	Close(ctx context.Context) error
}

type Neo4jKnowledgeStore struct {
	driver neo4j.DriverWithContext
	ollama *OllamaClient
}

func NewNeo4jKnowledgeStore(driver neo4j.DriverWithContext, ollama *OllamaClient) *Neo4jKnowledgeStore {
	return &Neo4jKnowledgeStore{driver: driver, ollama: ollama}
}

func (s *Neo4jKnowledgeStore) EnsureSchema(ctx context.Context) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	statements := []string{
		"CREATE CONSTRAINT entity_uuid IF NOT EXISTS FOR (e:Entity) REQUIRE e.uuid IS UNIQUE",
		"CREATE CONSTRAINT episode_uuid IF NOT EXISTS FOR (e:Episode) REQUIRE e.uuid IS UNIQUE",
		"CREATE VECTOR INDEX entity_embedding IF NOT EXISTS FOR (e:Entity) ON (e.embedding) OPTIONS {indexConfig: {`vector.dimensions`: 768, `vector.similarity_function`: 'cosine'}}",
		"CREATE FULLTEXT INDEX entity_name_ft IF NOT EXISTS FOR (e:Entity) ON EACH [e.name, e.summary]",
	}
	for _, stmt := range statements {
		if _, err := session.Run(ctx, stmt, nil); err != nil {
			log.Printf("schema statement warning: %v (statement: %s)", err, stmt)
		}
	}
	return nil
}

func (s *Neo4jKnowledgeStore) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

func (s *Neo4jKnowledgeStore) AddEpisode(ctx context.Context, req model.EpisodeRequest) (*model.EpisodeResponse, error) {
	entities, relations, err := s.ollama.ExtractEntities(ctx, req.Body)
	if err != nil {
		return nil, fmt.Errorf("extract entities: %w", err)
	}

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	var responseEntities []model.Entity
	entityUUIDs := make(map[string]string) // name -> uuid

	// MERGE entities and generate embeddings
	for _, e := range entities {
		entUUID := uuid.New().String()

		embedding, embedErr := s.ollama.Embed(ctx, e.Name+" "+e.Summary)
		if embedErr != nil {
			log.Printf("embedding error for entity %s: %v", e.Name, embedErr)
		}

		var params map[string]any
		if embedding != nil {
			params = map[string]any{
				"uuid":      entUUID,
				"name":      e.Name,
				"type":      e.Type,
				"summary":   e.Summary,
				"embedding": embedding,
			}
		} else {
			params = map[string]any{
				"uuid":    entUUID,
				"name":    e.Name,
				"type":    e.Type,
				"summary": e.Summary,
			}
		}

		cypher := `MERGE (e:Entity {name: $name, type: $type})
ON CREATE SET e.uuid = $uuid, e.summary = $summary
ON MATCH SET e.summary = $summary`
		if embedding != nil {
			cypher += `, e.embedding = $embedding`
		}
		cypher += ` RETURN e.uuid AS uuid`

		result, err := session.Run(ctx, cypher, params)
		if err != nil {
			log.Printf("merge entity error: %v", err)
			continue
		}
		if result.Next(ctx) {
			entUUID = result.Record().Values[0].(string)
		}

		entityUUIDs[e.Name] = entUUID
		responseEntities = append(responseEntities, model.Entity{
			UUID:    entUUID,
			Name:    e.Name,
			Type:    e.Type,
			Summary: e.Summary,
		})
	}

	// Create relations
	var responseRelations []model.Relation
	for _, r := range relations {
		srcUUID, srcOK := entityUUIDs[r.Source]
		tgtUUID, tgtOK := entityUUIDs[r.Target]
		if !srcOK || !tgtOK {
			continue
		}

		relUUID := uuid.New().String()
		_, err := session.Run(ctx,
			`MATCH (a:Entity {uuid: $src}), (b:Entity {uuid: $tgt})
			 MERGE (a)-[r:RELATION {type: $type}]->(b)
			 ON CREATE SET r.uuid = $uuid, r.fact = $fact
			 ON MATCH SET r.fact = $fact`,
			map[string]any{
				"src":  srcUUID,
				"tgt":  tgtUUID,
				"type": r.Type,
				"fact": r.Fact,
				"uuid": relUUID,
			},
		)
		if err != nil {
			log.Printf("create relation error: %v", err)
			continue
		}

		responseRelations = append(responseRelations, model.Relation{
			UUID:       relUUID,
			SourceUUID: srcUUID,
			TargetUUID: tgtUUID,
			Type:       r.Type,
			Fact:       r.Fact,
		})
	}

	// Create episode and MENTIONS edges
	episodeUUID := uuid.New().String()
	_, err = session.Run(ctx,
		`CREATE (ep:Episode {uuid: $uuid, name: $name, body: $body, source: $source, group_id: $group_id, created_at: datetime()})`,
		map[string]any{
			"uuid":     episodeUUID,
			"name":     req.Name,
			"body":     req.Body,
			"source":   req.Source,
			"group_id": req.GroupID,
		},
	)
	if err != nil {
		log.Printf("create episode error: %v", err)
	}

	// Link episode to entities
	for _, entUUID := range entityUUIDs {
		_, err = session.Run(ctx,
			`MATCH (ep:Episode {uuid: $ep_uuid}), (e:Entity {uuid: $ent_uuid})
			 CREATE (ep)-[:MENTIONS]->(e)`,
			map[string]any{"ep_uuid": episodeUUID, "ent_uuid": entUUID},
		)
		if err != nil {
			log.Printf("create mentions edge error: %v", err)
		}
	}

	return &model.EpisodeResponse{
		UUID:      episodeUUID,
		Name:      req.Name,
		Entities:  responseEntities,
		Relations: responseRelations,
	}, nil
}

func (s *Neo4jKnowledgeStore) SearchFacts(ctx context.Context, query string, groupID string) (*model.SearchFactsResponse, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	embedding, err := s.ollama.Embed(ctx, query)
	if err != nil {
		return &model.SearchFactsResponse{Facts: []model.Fact{}}, fmt.Errorf("embed query: %w", err)
	}

	var cypher string
	params := map[string]any{
		"embedding": embedding,
		"top_k":     20,
	}

	if groupID != "" {
		cypher = `CALL db.index.vector.queryNodes('entity_embedding', $top_k, $embedding)
YIELD node, score
MATCH (node)-[r:RELATION]-(other:Entity)
WHERE EXISTS {
  MATCH (ep:Episode {group_id: $group_id})-[:MENTIONS]->(node)
}
RETURN r.uuid AS uuid, r.type AS name, r.fact AS fact,
       node.uuid AS src_uuid, node.name AS src_name, node.type AS src_type, node.summary AS src_summary,
       other.uuid AS tgt_uuid, other.name AS tgt_name, other.type AS tgt_type, other.summary AS tgt_summary,
       score
ORDER BY score DESC`
		params["group_id"] = groupID
	} else {
		cypher = `CALL db.index.vector.queryNodes('entity_embedding', $top_k, $embedding)
YIELD node, score
MATCH (node)-[r:RELATION]-(other:Entity)
RETURN r.uuid AS uuid, r.type AS name, r.fact AS fact,
       node.uuid AS src_uuid, node.name AS src_name, node.type AS src_type, node.summary AS src_summary,
       other.uuid AS tgt_uuid, other.name AS tgt_name, other.type AS tgt_type, other.summary AS tgt_summary,
       score
ORDER BY score DESC`
	}

	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		return &model.SearchFactsResponse{Facts: []model.Fact{}}, fmt.Errorf("search facts query: %w", err)
	}

	var facts []model.Fact
	seen := make(map[string]bool)
	for result.Next(ctx) {
		rec := result.Record()
		factUUID, _ := rec.Get("uuid")
		uid := fmt.Sprintf("%v", factUUID)
		if seen[uid] {
			continue
		}
		seen[uid] = true

		name, _ := rec.Get("name")
		fact, _ := rec.Get("fact")
		srcUUID, _ := rec.Get("src_uuid")
		srcName, _ := rec.Get("src_name")
		srcType, _ := rec.Get("src_type")
		srcSummary, _ := rec.Get("src_summary")
		tgtUUID, _ := rec.Get("tgt_uuid")
		tgtName, _ := rec.Get("tgt_name")
		tgtType, _ := rec.Get("tgt_type")
		tgtSummary, _ := rec.Get("tgt_summary")

		facts = append(facts, model.Fact{
			UUID: toString(uid),
			Name: toString(name),
			Fact: toString(fact),
			SourceNode: model.Entity{
				UUID:    toString(srcUUID),
				Name:    toString(srcName),
				Type:    toString(srcType),
				Summary: toString(srcSummary),
			},
			TargetNode: model.Entity{
				UUID:    toString(tgtUUID),
				Name:    toString(tgtName),
				Type:    toString(tgtType),
				Summary: toString(tgtSummary),
			},
		})
	}

	if facts == nil {
		facts = []model.Fact{}
	}
	return &model.SearchFactsResponse{Facts: facts}, nil
}

func (s *Neo4jKnowledgeStore) GetGraph(ctx context.Context, limit int) (*model.GraphData, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (a:Entity)-[r:RELATION]->(b:Entity)
		 RETURN a.uuid AS a_uuid, a.name AS a_name, a.type AS a_type, a.summary AS a_summary,
		        r.uuid AS r_uuid, r.type AS r_type, r.fact AS r_fact,
		        b.uuid AS b_uuid, b.name AS b_name, b.type AS b_type, b.summary AS b_summary
		 LIMIT $limit`,
		map[string]any{"limit": limit},
	)
	if err != nil {
		return nil, fmt.Errorf("get graph query: %w", err)
	}

	nodeMap := make(map[string]model.GraphNode)
	var edges []model.GraphEdge

	for result.Next(ctx) {
		rec := result.Record()
		aUUID, _ := rec.Get("a_uuid")
		aName, _ := rec.Get("a_name")
		aType, _ := rec.Get("a_type")
		aSummary, _ := rec.Get("a_summary")
		rUUID, _ := rec.Get("r_uuid")
		rType, _ := rec.Get("r_type")
		rFact, _ := rec.Get("r_fact")
		bUUID, _ := rec.Get("b_uuid")
		bName, _ := rec.Get("b_name")
		bType, _ := rec.Get("b_type")
		bSummary, _ := rec.Get("b_summary")

		aID := toString(aUUID)
		bID := toString(bUUID)

		if _, ok := nodeMap[aID]; !ok {
			nodeMap[aID] = model.GraphNode{
				ID:      aID,
				Name:    toString(aName),
				Type:    toString(aType),
				Summary: toString(aSummary),
			}
		}
		if _, ok := nodeMap[bID]; !ok {
			nodeMap[bID] = model.GraphNode{
				ID:      bID,
				Name:    toString(bName),
				Type:    toString(bType),
				Summary: toString(bSummary),
			}
		}

		edges = append(edges, model.GraphEdge{
			ID:     toString(rUUID),
			Source: aID,
			Target: bID,
			Type:   toString(rType),
			Fact:   toString(rFact),
			Weight: 1.0,
		})
	}

	nodes := make([]model.GraphNode, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}
	if edges == nil {
		edges = []model.GraphEdge{}
	}

	return &model.GraphData{Nodes: nodes, Edges: edges}, nil
}

func (s *Neo4jKnowledgeStore) GetNode(ctx context.Context, id string, depth int) (*model.NodeDetail, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Get the target node
	nodeResult, err := session.Run(ctx,
		`MATCH (e:Entity {uuid: $id}) RETURN e.uuid AS uuid, e.name AS name, e.type AS type, e.summary AS summary`,
		map[string]any{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("get node query: %w", err)
	}
	if !nodeResult.Next(ctx) {
		return nil, fmt.Errorf("node not found: %s", id)
	}

	rec := nodeResult.Record()
	nUUID, _ := rec.Get("uuid")
	nName, _ := rec.Get("name")
	nType, _ := rec.Get("type")
	nSummary, _ := rec.Get("summary")

	node := model.GraphNode{
		ID:      toString(nUUID),
		Name:    toString(nName),
		Type:    toString(nType),
		Summary: toString(nSummary),
	}

	// Get neighbors and edges
	relResult, err := session.Run(ctx,
		`MATCH (e:Entity {uuid: $id})-[r:RELATION]-(n:Entity)
		 RETURN r.uuid AS r_uuid, r.type AS r_type, r.fact AS r_fact,
		        n.uuid AS n_uuid, n.name AS n_name, n.type AS n_type, n.summary AS n_summary,
		        startNode(r) = e AS is_outgoing`,
		map[string]any{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("get node relations: %w", err)
	}

	var neighbors []model.GraphNode
	var edges []model.GraphEdge
	seen := make(map[string]bool)

	for relResult.Next(ctx) {
		rec := relResult.Record()
		rUUID, _ := rec.Get("r_uuid")
		rType, _ := rec.Get("r_type")
		rFact, _ := rec.Get("r_fact")
		neighborUUID, _ := rec.Get("n_uuid")
		neighborName, _ := rec.Get("n_name")
		neighborType, _ := rec.Get("n_type")
		neighborSummary, _ := rec.Get("n_summary")
		isOutgoing, _ := rec.Get("is_outgoing")

		nID := toString(neighborUUID)
		if !seen[nID] {
			seen[nID] = true
			neighbors = append(neighbors, model.GraphNode{
				ID:      nID,
				Name:    toString(neighborName),
				Type:    toString(neighborType),
				Summary: toString(neighborSummary),
			})
		}

		var src, tgt string
		if isOutgoing.(bool) {
			src = id
			tgt = nID
		} else {
			src = nID
			tgt = id
		}

		edges = append(edges, model.GraphEdge{
			ID:     toString(rUUID),
			Source: src,
			Target: tgt,
			Type:   toString(rType),
			Fact:   toString(rFact),
			Weight: 1.0,
		})
	}

	if neighbors == nil {
		neighbors = []model.GraphNode{}
	}
	if edges == nil {
		edges = []model.GraphEdge{}
	}

	return &model.NodeDetail{
		Node:      node,
		Neighbors: neighbors,
		Edges:     edges,
	}, nil
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
