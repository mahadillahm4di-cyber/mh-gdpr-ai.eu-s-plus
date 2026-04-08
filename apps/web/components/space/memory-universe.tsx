"use client";

import { useMemo } from "react";
import { Canvas } from "@react-three/fiber";
import { OrbitControls, Stars } from "@react-three/drei";
import {
  EffectComposer,
  Bloom,
  Vignette,
  ChromaticAberration,
} from "@react-three/postprocessing";
import { BlendFunction } from "postprocessing";
import { Vector2 } from "three";
import { MemoryStar } from "./memory-star";
import { StarConnections } from "./star-connections";
import { DustParticles } from "./dust-particles";
import type { Memory } from "@/lib/api";

interface MemoryUniverseProps {
  memories: Memory[];
  onSelectMemory: (memory: Memory) => void;
}

export function MemoryUniverse({
  memories,
  onSelectMemory,
}: MemoryUniverseProps) {
  const chromaticOffset = useMemo(() => new Vector2(0.002, 0.002), []);

  return (
    <Canvas
      camera={{ position: [0, 0, 8], fov: 60 }}
      style={{ background: "#000000" }}
    >
      {/* Lighting */}
      <ambientLight intensity={0.3} />
      <pointLight position={[10, 10, 10]} intensity={0.5} />

      {/* Background stars */}
      <Stars
        radius={100}
        depth={50}
        count={5000}
        factor={4}
        saturation={0}
        fade
        speed={0.5}
      />

      {/* Dust particles */}
      <DustParticles count={200} />

      {/* Memory stars with staggered entry animation */}
      {memories.map((memory, index) => (
        <MemoryStar
          key={memory.id}
          memory={memory}
          onClick={onSelectMemory}
          delay={index * 0.1}
        />
      ))}

      {/* Connections between related memories */}
      <StarConnections memories={memories} />

      {/* Camera controls */}
      <OrbitControls
        enablePan
        enableZoom
        enableRotate
        autoRotate
        autoRotateSpeed={0.3}
        maxDistance={20}
        minDistance={2}
        dampingFactor={0.05}
        enableDamping
      />

      {/* Post-processing effects */}
      <EffectComposer>
        <Bloom
          luminanceThreshold={0.2}
          luminanceSmoothing={0.9}
          intensity={1.5}
        />
        <ChromaticAberration
          blendFunction={BlendFunction.NORMAL}
          offset={chromaticOffset}
          radialModulation={false}
          modulationOffset={0}
        />
        <Vignette eskil={false} offset={0.1} darkness={0.8} />
      </EffectComposer>
    </Canvas>
  );
}
