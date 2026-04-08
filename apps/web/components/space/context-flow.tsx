"use client";

import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import type { Points } from "three";

interface ContextFlowProps {
  from: [number, number, number];
  to: [number, number, number];
  color: string;
  active: boolean;
}

/**
 * ContextFlow creates an animated particle stream between two points.
 * Used to visualize context flowing between providers during a switch.
 */
export function ContextFlow({ from, to, color, active }: ContextFlowProps) {
  const pointsRef = useRef<Points>(null);
  const particleCount = 30;

  // Generate particle positions along the path
  const positions = new Float32Array(particleCount * 3);
  for (let i = 0; i < particleCount; i++) {
    const t = i / particleCount;
    positions[i * 3] = from[0] + (to[0] - from[0]) * t;
    positions[i * 3 + 1] = from[1] + (to[1] - from[1]) * t;
    positions[i * 3 + 2] = from[2] + (to[2] - from[2]) * t;
  }

  // Animate particles along the path
  useFrame((state) => {
    if (!pointsRef.current || !active) return;
    const geo = pointsRef.current.geometry;
    const pos = geo.attributes.position;
    const t = state.clock.elapsedTime;

    for (let i = 0; i < particleCount; i++) {
      const progress = ((i / particleCount + t * 0.5) % 1);
      pos.setXYZ(
        i,
        from[0] + (to[0] - from[0]) * progress + Math.sin(t * 2 + i) * 0.05,
        from[1] + (to[1] - from[1]) * progress + Math.cos(t * 2 + i) * 0.05,
        from[2] + (to[2] - from[2]) * progress
      );
    }
    pos.needsUpdate = true;
  });

  if (!active) return null;

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          args={[positions, 3]}
        />
      </bufferGeometry>
      <pointsMaterial
        color={color}
        size={0.04}
        transparent
        opacity={0.8}
        sizeAttenuation
      />
    </points>
  );
}
