import React, { useRef, useState } from "react";
import NodeCard from "./NodeCard";

export default function CanvasArea() {
  const [nodes, setNodes] = useState([]);
  const canvasRef = useRef(null);

  function allowDrop(e) {
    e.preventDefault();
  }

  function handleDrop(e) {
    e.preventDefault();
    const payload = e.dataTransfer.getData("autoflow-item");
    if (!payload) return;
    const item = JSON.parse(payload);
    const rect = canvasRef.current.getBoundingClientRect();
    const x = Math.round(e.clientX - rect.left);
    const y = Math.round(e.clientY - rect.top);
    const id = `n_${Date.now()}`;
    const newNode = { id, type: item.id, label: item.label, x, y, width: 160, height: 64 };
    setNodes((s) => [...s, newNode]);
  }

  function removeNode(id) {
    setNodes((s) => s.filter((n) => n.id !== id));
  }

  return (
    <div className="h-[76vh] bg-white rounded-2xl shadow p-3 flex flex-col">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-lg font-semibold">Workflow Canvas</h3>
        <div className="text-sm text-gray-500">Drag components onto canvas</div>
      </div>

      <div
        ref={canvasRef}
        onDragOver={allowDrop}
        onDrop={handleDrop}
        className="relative flex-1 border border-dashed border-gray-200 rounded-lg overflow-hidden bg-gradient-to-br from-white to-gray-50"
      >
        <div className="absolute inset-0 bg-[repeating-linear-gradient(0deg,#e6e6e6_0_1px,transparent_1px,transparent_20px),repeating-linear-gradient(90deg,#e6e6e6_0_1px,transparent_1px,transparent_20px)] opacity-30 pointer-events-none" />

        {nodes.map((n) => (
          <NodeCard key={n.id} node={n} onDelete={() => removeNode(n.id)} />
        ))}
      </div>
    </div>
  );
}
