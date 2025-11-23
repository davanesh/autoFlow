import React from "react";

export default function PropertiesPanel({ node, connections, onUpdateConnection, onUpdateNode }) {
  if (!node) return null; // ✅ Prevent crash if node is deleted

  const relatedConnections = connections.filter((c) => c.from === node.id || c.to === node.id);

  const handleNodeChange = (field, value) => {
    onUpdateNode({ ...node, [field]: value });
  };

  return (
    <div className="w-72 p-4 border rounded-lg bg-white shadow flex-shrink-0 space-y-4">
      <h4 className="text-lg font-semibold mb-2">Node Properties</h4>

      {/* Node Label */}
      <div className="flex flex-col">
        <label className="text-sm font-medium mb-1">Label</label>
        <input
          className="w-full border p-1 rounded"
          value={node.label}
          onChange={(e) => handleNodeChange("label", e.target.value)}
        />
      </div>

      {/* Node Color */}
      <div className="flex flex-col">
        <label className="text-sm font-medium mb-1">Color</label>
        <input
          type="color"
          className="w-full h-8 cursor-pointer border rounded"
          value={node.color}
          onChange={(e) => handleNodeChange("color", e.target.value)}
        />
      </div>

      {/* Node Size */}
      <div className="flex flex-col space-y-1">
        <label className="text-sm font-medium mb-1">Size</label>
        <div className="flex gap-2">
          <input
            type="number"
            className="w-1/2 border p-1 rounded"
            value={node.width}
            onChange={(e) => handleNodeChange("width", parseInt(e.target.value))}
          />
          <input
            type="number"
            className="w-1/2 border p-1 rounded"
            value={node.height}
            onChange={(e) => handleNodeChange("height", parseInt(e.target.value))}
          />
        </div>
      </div>

      {/* Connection Labels */}
      {relatedConnections.length > 0 && (
        <div className="space-y-2">
          <h5 className="text-sm font-medium">Connections</h5>
          {relatedConnections.map((c) => (
            <input
              key={c.from + "-" + c.to}
              className="w-full border p-1 rounded text-sm"
              placeholder="Connection Label"
              value={c.label || ""}
              onChange={(e) => onUpdateConnection({ ...c, label: e.target.value })}
            />
          ))}
        </div>
      )}
      {/* AI NODE FIELDS */}
    {node.type === "ai" && (
      <div className="space-y-2 mt-3">
        
        <label className="text-xs font-semibold text-gray-600">AI Prompt</label>
        <textarea
          className="w-full border rounded p-2 text-sm"
          value={node.data?.prompt || ""}
          onChange={(e) =>
            onUpdateNode({
              ...node,
              data: { ...node.data, prompt: e.target.value }
            })
          }
        />

        <label className="text-xs font-semibold text-gray-600">AI Input</label>
        <textarea
          className="w-full border rounded p-2 text-sm"
          value={node.data?.input || ""}
          onChange={(e) =>
            onUpdateNode({
              ...node,
              data: { ...node.data, input: e.target.value }
            })
          }
        />
      </div>
    )}
    </div>
  );
}
