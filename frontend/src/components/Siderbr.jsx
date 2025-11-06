import React from "react";

const TOOL_ITEMS = [
  { id: "task", label: "Task" },
  { id: "decision", label: "Decision" },
  { id: "start", label: "Start" },
  { id: "end", label: "End" },
];

export default function Sidebar() {
  return (
    <div className="bg-white p-4 rounded-2xl shadow">
      <h3 className="font-semibold mb-3">Components</h3>

      <div className="space-y-2">
        {TOOL_ITEMS.map((t) => (
          <div
            key={t.id}
            draggable
            onDragStart={(e) => e.dataTransfer.setData("autoflow-item", JSON.stringify(t))}
            className="p-3 rounded-lg border border-dashed border-gray-200 bg-gray-50 cursor-move"
          >
            <div className="font-medium">{t.label}</div>
            <div className="text-xs text-gray-400">{t.id}</div>
          </div>
        ))}
      </div>

      <div className="mt-6 border-t pt-4">
        <button className="w-full py-2 bg-indigo-600 text-white rounded-lg">Save</button>
        <button className="w-full py-2 border rounded-lg mt-2">Export</button>
      </div>
    </div>
  );
}
