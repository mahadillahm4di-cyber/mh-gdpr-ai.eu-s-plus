"use client";

import { useRef, useState } from "react";
import { useFrame } from "@react-three/fiber";
import { Html } from "@react-three/drei";
import type { Mesh } from "three";
import type { Memory } from "@/lib/api";

const PROVIDER_COLORS: Record<string, string> = {
  openai: "#4A90D9",
  anthropic: "#E87B35",
  ollama: "#4ADE80",
};

interface MemoryStarProps {
  memory: Memory;
  onClick: (memory: Memory) => void;
  delay?: number;
}

export function MemoryStar({ memory, onClick, delay = 0 }: MemoryStarProps) {
  const meshRef = useRef<Mesh>(null);
  const [hovered, setHovered] = useState(false);
  const [visible, setVisible] = useState(false);

  const color = PROVIDER_COLORS[memory.theme] || "#8B5CF6";
  const size = 0.1 + memory.importance * 0.3;

  useFrame((state) => {
    if (!meshRef.current) return;
    const t = state.clock.elapsedTime;

    // Staggered entry animation
    if (!visible && t > delay) {
      setVisible(true);
    }

    if (!visible) {
      meshRef.current.scale.setScalar(0);
      return;
    }

    // Smooth scale-in
    const entryProgress = Math.min((t - delay) / 0.5, 1);
    const eased = 1 - Math.pow(1 - entryProgress, 3); // easeOutCubic

    // Floating animation
    meshRef.current.position.y +=
      Math.sin(t * 0.5 + memory.position_x) * 0.0005;
    meshRef.current.position.x +=
      Math.cos(t * 0.3 + memory.position_y) * 0.0003;

    // Scale with hover pulse
    if (hovered) {
      const pulseScale = 1.3 + Math.sin(t * 3) * 0.1;
      meshRef.current.scale.setScalar(pulseScale * eased);
    } else {
      meshRef.current.scale.setScalar(eased);
    }
  });

  return (
    <group
      position={[memory.position_x, memory.position_y, memory.position_z]}
    >
      <mesh
        ref={meshRef}
        onClick={() => onClick(memory)}
        onPointerOver={() => setHovered(true)}
        onPointerOut={() => setHovered(false)}
      >
        <sphereGeometry args={[size, 32, 32]} />
        <meshStandardMaterial
          color={color}
          emissive={color}
          emissiveIntensity={hovered ? 2 : 0.8}
          transparent
          opacity={0.9}
        />
      </mesh>

      {/* Glow effect */}
      <mesh>
        <sphereGeometry args={[size * 2, 16, 16]} />
        <meshBasicMaterial
          color={color}
          transparent
          opacity={hovered ? 0.15 : 0.05}
        />
      </mesh>

      {/* Tooltip on hover */}
      {hovered && (
        <Html distanceFactor={10} center>
          <div className="pointer-events-none max-w-[200px] rounded-lg border border-white/20 bg-black/90 px-3 py-2 text-xs text-white/80 shadow-xl backdrop-blur-sm">
            <div className="mb-1 font-semibold text-white">
              {memory.theme || "Memory"}
            </div>
            <div className="line-clamp-3">{memory.summary}</div>
          </div>
        </Html>
      )}
    </group>
  );
}
