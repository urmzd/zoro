pub mod agent;
pub mod commands;
pub mod config;
pub mod event_store;
pub mod knowledge;
pub mod models;
pub mod ollama;
pub mod orchestrator;
pub mod searcher;
pub mod tools;

use commands::AppState;
use std::sync::Arc;
use tauri::Manager;
use tauri_plugin_shell::ShellExt;

pub fn run() {
    env_logger::init();

    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let cfg = config::AppConfig::default();

            // Kill any orphaned SearXNG process on port 8888
            if std::net::TcpStream::connect("127.0.0.1:8888").is_ok() {
                log::warn!("Port 8888 already in use, killing existing process");
                #[cfg(unix)]
                {
                    let _ = std::process::Command::new("sh")
                        .args(["-c", "lsof -ti :8888 | xargs kill 2>/dev/null"])
                        .status();
                    std::thread::sleep(std::time::Duration::from_secs(1));
                }
            }

            // Spawn SearXNG sidecar
            let sidecar_command = app.shell().sidecar("searxng").unwrap();
            let (mut rx, child) = sidecar_command.spawn().expect("failed to spawn searxng sidecar");

            // Log sidecar output in background
            tauri::async_runtime::spawn(async move {
                use tauri_plugin_shell::process::CommandEvent;
                while let Some(event) = rx.recv().await {
                    match event {
                        CommandEvent::Stdout(line) => {
                            log::info!("[searxng] {}", String::from_utf8_lossy(&line));
                        }
                        CommandEvent::Stderr(line) => {
                            log::warn!("[searxng] {}", String::from_utf8_lossy(&line));
                        }
                        CommandEvent::Terminated(status) => {
                            log::info!("[searxng] terminated: {:?}", status);
                            break;
                        }
                        _ => {}
                    }
                }
            });

            // Store child handle for cleanup on exit
            app.manage(SidecarChild(std::sync::Mutex::new(Some(child))));

            // Wait for SearXNG to become ready (poll TCP port)
            for i in 0..30 {
                if std::net::TcpStream::connect("127.0.0.1:8888").is_ok() {
                    log::info!("SearXNG sidecar is ready (attempt {})", i + 1);
                    break;
                }
                if i == 29 {
                    log::warn!("SearXNG sidecar did not become ready within 15s");
                }
                std::thread::sleep(std::time::Duration::from_millis(500));
            }

            let knowledge = tauri::async_runtime::block_on(async {
                let ks = knowledge::KnowledgeStore::new(&cfg.db_path)
                    .await
                    .expect("open knowledge store");
                ks.ensure_schema().await.ok();
                ks
            });

            let knowledge = Arc::new(knowledge);

            let event_store = event_store::EventStore::new(
                knowledge.db().clone(),
            );
            let event_store = Arc::new(tauri::async_runtime::block_on(async {
                event_store.ensure_schema().await.ok();
                event_store
            }));

            let ollama = Arc::new(ollama::OllamaClient::new(
                &cfg.ollama_host,
                &cfg.ollama_model,
                &cfg.embedding_model,
            ));

            let searcher = Arc::new(searcher::Searcher::new());

            let registry = Arc::new(models::ModelRegistry::new(
                cfg.ollama_model.clone(),
                cfg.ollama_fast_model.clone(),
                cfg.embedding_model.clone(),
            ));

            let tool_registry = Arc::new(tools::ToolRegistry::new(
                searcher.clone(),
                knowledge.clone(),
                ollama.clone(),
            ));

            let agent = Arc::new(agent::Agent::new(
                ollama.clone(),
                tool_registry.clone(),
                registry,
                event_store,
            ));

            let orchestrator = Arc::new(orchestrator::Orchestrator::new(
                knowledge.clone(),
                ollama.clone(),
                searcher,
            ));

            app.manage(AppState {
                agent,
                orchestrator,
                knowledge,
                ollama,
            });

            Ok(())
        })
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::Destroyed = event {
                if let Some(sidecar) = window.app_handle().try_state::<SidecarChild>() {
                    if let Some(child) = sidecar.0.lock().unwrap().take() {
                        log::info!("Killing SearXNG sidecar process");
                        let _ = child.kill();
                    }
                }
            }
        })
        .invoke_handler(tauri::generate_handler![
            commands::create_chat_session,
            commands::list_chat_sessions,
            commands::get_chat_session,
            commands::send_chat_message,
            commands::start_research,
            commands::search_knowledge,
            commands::get_knowledge_graph,
            commands::get_node_detail,
            commands::classify_intent,
            commands::get_autocomplete,
        ])
        .run(tauri::generate_context!())
        .expect("error running tauri application");
}

struct SidecarChild(std::sync::Mutex<Option<tauri_plugin_shell::process::CommandChild>>);
