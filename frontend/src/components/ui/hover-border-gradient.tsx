"use client";

import { motion } from "framer-motion";
import { useState } from "react";
import { cn } from "@/lib/utils";

interface HoverBorderGradientProps {
  children: React.ReactNode;
  containerClassName?: string;
  className?: string;
  as?: React.ElementType;
  duration?: number;
  [key: string]: unknown;
}

export function HoverBorderGradient({
  children,
  containerClassName,
  className,
  as: Component = "div",
  duration = 1,
  ...props
}: HoverBorderGradientProps) {
  const [hovered, setHovered] = useState(false);

  return (
    <Component
      className={cn("relative rounded-full p-[1px] group/btn", containerClassName)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      {...props}
    >
      <motion.div
        className="absolute inset-0 rounded-full opacity-0 group-hover/btn:opacity-100"
        style={{
          background: "linear-gradient(90deg, #6366f1, #a855f7, #ec4899, #6366f1)",
          backgroundSize: "300% 100%",
        }}
        animate={
          hovered ? { backgroundPosition: ["0% 0%", "100% 0%"] } : { backgroundPosition: "0% 0%" }
        }
        transition={{ duration, repeat: hovered ? Infinity : 0, ease: "linear" }}
      />
      <div className={cn("relative rounded-full bg-background flex items-center", className)}>
        {children}
      </div>
    </Component>
  );
}
