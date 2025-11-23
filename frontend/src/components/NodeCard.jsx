import { useRef } from "react";
import { ArrowRight } from "lucide-react";

const NodeCard = ({
  node,
  onMouseDown,
  onClick,
  onDelete,
  onStartConnection,
  isSelected,
}) => {
  const nodeRef = useRef(null);

  return (
    <div
      ref={nodeRef}
      onMouseDown={(e) => onMouseDown?.(e, node.id)}
      onClick={(e) => onClick?.(e, node.id)}
      style={{
        position: "absolute",
        left: `${node.x}px`,
        top: `${node.y}px`,
        width: `${node.width}px`,
        height: `${node.height}px`,
        userSelect: "none",
        backgroundColor:
        node.type === "ai"
        ? "#e8d5ff"     // light purple
        : node.color || "#fff",
        transition: "background-color 0.2s ease",
      }}
      className={`rounded-lg shadow-md border ${
        isSelected ? "ring-2 ring-blue-400" : "border-gray-200"
      }`}
    >
      <div className="relative p-3 w-full h-full cursor-grab active:cursor-grabbing">
        <h3 className="text-sm font-semibold text-gray-800 truncate">
          {node.label || "Untitled Node"}
        </h3>

        {/* Arrow connector */}
        <button
          onMouseDown={(e) => {
            e.stopPropagation();
            onStartConnection?.(node.id);
          }}
          className="absolute right-[-12px] top-1/2 -translate-y-1/2 bg-blue-500 text-white p-2 rounded-full shadow-md hover:bg-blue-600 active:scale-95 transition"
        >
          <ArrowRight size={14} />
        </button>

        {/* Delete button */}
        <button
          onMouseDown={(e) => e.stopPropagation()}
          onClick={(e) => {
            e.stopPropagation();
            onDelete?.(node.id);
          }}
          className="absolute -top-2 -left-2 w-6 h-6 bg-red-500 rounded-full flex items-center justify-center text-white text-sm font-bold hover:bg-red-600 transition"
        >
          ×
        </button>
      </div>
    </div>
  );
};

export default NodeCard;
