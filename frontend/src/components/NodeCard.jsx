import React from "react";

export default function NodeCard({ node, onDelete }) {
  // inline style for position
  const style = {
    left: node.x,
    top: node.y,
    width: node.width,
    height: node.height,
  };

  return (
    <div
      className="absolute p-2 select-none rounded-2xl shadow cursor-grab"
      style={style}
      onMouseDown={(e) => e.stopPropagation()}
    >
      <div className="w-full h-full flex items-center justify-center bg-white border border-gray-200 rounded-xl">
        <div className="text-sm font-semibold">{node.label}</div>
      </div>

      <button
        onClick={(ev) => {
          ev.stopPropagation();
          onDelete();
        }}
        className="absolute -top-3 -right-3 bg-red-500 text-white rounded-full w-6 h-6 text-xs flex items-center justify-center shadow"
        title="Delete"
      >
        ×
      </button>
    </div>
  );
}
