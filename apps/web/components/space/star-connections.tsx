"use client";

import { useMemo } from "react";
import { Line } from "@react-three/drei";
import type { Memory } from "@/lib/api";

interface StarConnectionsProps {
  memories: Memory[];
}

export function StarConnections({ memories }: StarConnectionsProps) {
  // Connect memories that share the same theme
  const connections = useMemo(() => {
    const links: { from: Memory; to: Memory }[] = [];
    for (let i = 0; i < memories.length; i++) {
      for (let j = i + 1; j < memories.length; j++) {
        if (
          memories[i].theme &&
          memories[i].theme === memories[j].theme
        ) {
          links.push({ from: memories[i], to: memories[j] });
        }
      }
    }
    return links;
  }, [memories]);

  return (
    <>
      {connections.map((link, i) => (
        <Line
          key={i}
          points={[
            [link.from.position_x, link.from.position_y, link.from.position_z],
            [link.to.position_x, link.to.position_y, link.to.position_z],
          ]}
          color="#ffffff"
          opacity={0.08}
          transparent
          lineWidth={0.5}
        />
      ))}
    </>
  );
}
