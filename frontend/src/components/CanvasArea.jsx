import React, { useRef, useState } from "react";
import axios from "axios";
import NodeCard from "./NodeCard";
import PropertiesPanel from "./PropertiesPanel";

const GRID_SIZE = 20;
const API_BASE = import.meta.env.VITE_API_BASE;

export default function CanvasArea() {
  const [nodes, setNodes] = useState([]);
  const [connections, setConnections] = useState([]);
  const [selectedNodeIds, setSelectedNodeIds] = useState([]);
  const [draggingNodeIds, setDraggingNodeIds] = useState([]);
  const [offsets, setOffsets] = useState({});
  const [lineStartNode, setLineStartNode] = useState(null);
  const [tempLine, setTempLine] = useState(null);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [panning, setPanning] = useState(false);
  const [panStart, setPanStart] = useState({});
  const [loadId, setLoadId] = useState("");
  const canvasRef = useRef(null);

  // -------------------- Canvas Drag & Drop --------------------
  function allowDrop(e) { e.preventDefault(); }

  function handleDrop(e) {
    e.preventDefault();
    const payload = e.dataTransfer.getData("autoflow-item");
    if (!payload) return;
    const item = JSON.parse(payload);
    const rect = canvasRef.current.getBoundingClientRect();
    const x = Math.round((e.clientX - rect.left - pan.x) / (GRID_SIZE * zoom)) * GRID_SIZE;
    const y = Math.round((e.clientY - rect.top - pan.y) / (GRID_SIZE * zoom)) * GRID_SIZE;
    const id = `n_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
    setNodes((prev) => [
      ...prev,
      { id, type: item.id, label: item.label, x, y, width: 160, height: 64, color: "#fff", data: item.data || {} }
    ]);
  }

  function startDrag(e, nodeId) {
    e.stopPropagation(); e.preventDefault();
    const selected = selectedNodeIds.includes(nodeId) ? selectedNodeIds : [nodeId];
    setDraggingNodeIds(selected);

    const rect = canvasRef.current.getBoundingClientRect();
    const newOffsets = {};
    selected.forEach((id) => {
      const n = nodes.find((node) => node.id === id);
      if (!n) return;
      newOffsets[id] = { x: (e.clientX - rect.left - pan.x) / zoom - n.x, y: (e.clientY - rect.top - pan.y) / zoom - n.y };
    });
    setOffsets(newOffsets);
  }

  function onMouseMove(e) {
    const rect = canvasRef.current.getBoundingClientRect();
    const mouseX = (e.clientX - rect.left - pan.x) / zoom;
    const mouseY = (e.clientY - rect.top - pan.y) / zoom;

    if (draggingNodeIds.length) {
      e.preventDefault();
      setNodes((prev) =>
        prev.map((n) => {
          if (!draggingNodeIds.includes(n.id)) return n;
          const offset = offsets[n.id]; if (!offset) return n;
          return { ...n, x: Math.round((mouseX - offset.x) / GRID_SIZE) * GRID_SIZE, y: Math.round((mouseY - offset.y) / GRID_SIZE) * GRID_SIZE };
        })
      );
      return;
    }

    if (lineStartNode) {
      const startNode = nodes.find((n) => n.id === lineStartNode);
      if (startNode) {
        const start = { x: startNode.x + startNode.width / 2, y: startNode.y + startNode.height / 2 };
        setTempLine({ x1: start.x, y1: start.y, x2: mouseX, y2: mouseY });
      }
    }

    if (panning) {
      setPan({ x: panStart.x + (e.clientX - panStart.startX), y: panStart.y + (e.clientY - panStart.startY) });
    }
  }

  function onMouseUp(e) {
    if (draggingNodeIds.length) { setDraggingNodeIds([]); return; }

    if (lineStartNode) {
      const rect = canvasRef.current.getBoundingClientRect();
      const mouseX = (e.clientX - rect.left - pan.x) / zoom;
      const mouseY = (e.clientY - rect.top - pan.y) / zoom;
      const targetNode = nodes.find(
        (n) => mouseX >= n.x && mouseX <= n.x + n.width && mouseY >= n.y && mouseY <= n.y + n.height && n.id !== lineStartNode
      );
      if (targetNode) setConnections((prev) => [...prev, { from: lineStartNode, to: targetNode.id, label: "" }]);
      setLineStartNode(null); setTempLine(null);
    }

    if (panning) setPanning(false);
  }

  function onWheel(e) { e.preventDefault(); const scaleAmount = e.deltaY < 0 ? 1.1 : 0.9; setZoom((prev) => Math.max(0.3, Math.min(2.5, prev * scaleAmount))); }
  function onMiddleMouseDown(e) { e.preventDefault(); setPanning(true); setPanStart({ x: pan.x, y: pan.y, startX: e.clientX, startY: e.clientY }); }

  function handleNodeClick(e, nodeId) {
    if (draggingNodeIds.length) return;
    e.stopPropagation();
    if (e.shiftKey) setSelectedNodeIds((prev) => prev.includes(nodeId) ? prev.filter((id) => id !== nodeId) : [...prev, nodeId]);
    else setSelectedNodeIds([nodeId]);
  }

  function removeNode(id) {
    setNodes((s) => s.filter((n) => n.id !== id));
    setConnections((c) => c.filter((conn) => conn.from !== id && conn.to !== id));
    setSelectedNodeIds((prev) => prev.filter((nid) => nid !== id));
  }

  const handleUpdateNode = (updatedNode) => setNodes((prevNodes) => prevNodes.map((node) => node.id === updatedNode.id ? updatedNode : node));
  const handleUpdateConnection = (updatedConnection) => setConnections((prevConnections) => prevConnections.map((conn) => conn.from === updatedConnection.from && conn.to === updatedConnection.to ? updatedConnection : conn));

  const selectedNode = selectedNodeIds.length > 0 ? nodes.find((n) => n.id === selectedNodeIds[0]) : null;

  // -------------------- SAVE / LOAD --------------------
  function buildWorkflowPayload({ name = "Untitled", description = "", nodesArr = nodes, conns = connections } = {}) {
    const mappedNodes = nodesArr.map((n) => ({
      canvasId: n.id || n.canvasId,
      type: n.type,
      label: n.label,
      position: { x: n.x, y: n.y },
      width: n.width,
      height: n.height,
      data: n.data || {},
      lambdaName: n.lambdaName || (n.data?.lambdaName) || "",
      status: n.status || "draft",
    }));

    const mappedConns = conns.map((c) => ({
      source: c.from,
      target: c.to,
      label: c.label || "",
    }));

    return { name, description, nodes: mappedNodes, connections: mappedConns };
  }

  async function saveWorkflow() {
    try {
      const payload = buildWorkflowPayload();
      const res = await axios.post(`${API_BASE}/workflows`, payload);
      console.log("Saved workflow:", res.data);
      alert("Workflow saved! id: " + (res.data.id || res.data._id || ""));
    } catch (err) {
      console.error("Save failed:", err.response?.data || err.message);
      alert("Save failed: " + (err.response?.data?.error || err.message));
    }
  }

  async function loadWorkflowById(id) {
    if (!id) return alert("Enter workflow id to load in the input box");
    try {
      const res = await axios.get(`${API_BASE}/workflows`);
      const wfList = res.data || [];

      const getId = (w) => w._id?.$oid || w._id?.toString() || w.id;
      const wf = wfList.find((w) => getId(w) === id);

      if (!wf) return alert("Workflow not found. Check ID.");
      populateFromWorkflow(wf);
    } catch (err) {
      console.error("Load failed:", err.response?.data || err.message);
      alert("Load failed: " + (err.response?.data?.error || err.message));
    }
  }

function populateFromWorkflow(wf) {
  if (!wf) return;

  // Step 1: Normalize nodes
  const loadedNodes = (wf.nodes || []).map((n) => {
    const id = n.id || n.canvasId || `n_${Date.now()}_${Math.random().toString(36).substr(2,5)}`;
    return {
      ...n,
      id,
      canvasId: id,
      x: n.position?.x || 0,
      y: n.position?.y || 0,
      width: n.width || 160,
      height: n.height || 64,
      data: n.data || {},
      lambdaName: n.lambdaName || n.data?.lambdaName || "",
      status: n.status || "draft",
    };
  });

  // Step 2: Build a map of old IDs -> new IDs
  const idMap = {};
  (wf.nodes || []).forEach((n, idx) => {
    const normalizedId = loadedNodes[idx].id;
    idMap[n.id || n.canvasId] = normalizedId;
  });

  // Step 3: Remap connections using idMap
  const loadedConns = (wf.connections || [])
    .map((c) => {
      const from = idMap[c.source];
      const to = idMap[c.target];
      if (!from || !to) {
        console.warn("Connection render skipped — node missing:", c);
        return null;
      }
      return { from, to, label: c.label || "" };
    })
    .filter(Boolean);

  setNodes(loadedNodes);
  setConnections(loadedConns);
  setSelectedNodeIds([]);
  alert("Workflow loaded: " + (wf.name || wf._id || wf.id || ""));
}


  async function runWorkflowById(id) {
    if (!id) return alert("Enter workflow id to run in the input box");
    try {
      const res = await axios.post(`${API_BASE}/workflows/${id}/run`, {});
      console.log("Run response:", res.data);
      alert("Run finished — check console or Execution Logs UI.");
    } catch (err) {
      console.error("Run failed:", err.response?.data || err.message);
      alert("Run failed: " + (err.response?.data?.error || err.message));
    }
  }

  // -------------------- RENDER --------------------
  return (
    <div className="flex gap-4 select-none">
      <div
        className="h-[76vh] flex-1 bg-white rounded-2xl shadow p-3 flex flex-col relative overflow-hidden"
        ref={canvasRef}
        onMouseMove={onMouseMove}
        onMouseUp={onMouseUp}
        onClick={() => setSelectedNodeIds([])}
        onWheel={onWheel}
        onMouseDown={(e) => e.button === 1 && onMiddleMouseDown(e)}
      >
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-lg font-semibold">Workflow Canvas</h3>
          <div className="flex items-center gap-2">
            <input className="border rounded px-2 py-1 text-sm" placeholder="workflow id (load/run)" value={loadId} onChange={(e) => setLoadId(e.target.value)} />
            <button onClick={() => loadWorkflowById(loadId)} className="px-3 py-1 bg-gray-200 rounded hover:bg-gray-300 text-sm">Load</button>
            <button onClick={saveWorkflow} className="px-3 py-1 bg-green-500 rounded text-white hover:brightness-105 text-sm">Save</button>
            <button onClick={() => runWorkflowById(loadId)} className="px-3 py-1 bg-blue-500 rounded text-white hover:brightness-105 text-sm">Run</button>
          </div>
        </div>

        <div onDragOver={allowDrop} onDrop={handleDrop} className="relative flex-1 border border-dashed border-gray-200 rounded-lg overflow-hidden bg-gradient-to-br from-white to-gray-50">
          <div className="absolute inset-0 bg-[repeating-linear-gradient(0deg,#e6e6e6_0_1px,transparent_1px,transparent_20px),repeating-linear-gradient(90deg,#e6e6e6_0_1px,transparent_1px,transparent_20px)] opacity-30 pointer-events-none" style={{ transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`, transformOrigin: "top left" }} />

          <svg className="absolute inset-0 w-full h-full pointer-events-none" style={{ transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`, transformOrigin: "top left" }}>
            <defs><marker id="arrowhead" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto"><polygon points="0 0, 10 3.5, 0 7" fill="black" /></marker></defs>
            {connections.map((conn, idx) => {
              const fromNode = nodes.find((n) => n.id === conn.from);
              const toNode = nodes.find((n) => n.id === conn.to);
              if (!fromNode || !toNode) {
                console.warn("Connection render skipped — node missing:", conn);
                return null;
              }
              const start = { x: fromNode.x + fromNode.width / 2, y: fromNode.y + fromNode.height / 2 };
              const end = { x: toNode.x + toNode.width / 2, y: toNode.y + toNode.height / 2 };
              const deltaX = Math.abs(end.x - start.x) / 2;
              const pathD = `M${start.x},${start.y} C${start.x + deltaX},${start.y} ${end.x - deltaX},${end.y} ${end.x},${end.y}`;
              const midX = (start.x + end.x) / 2;
              const midY = (start.y + end.y) / 2;
              return (
                <g key={idx}>
                  <path d={pathD} stroke="black" strokeWidth="2" fill="none" markerEnd="url(#arrowhead)" />
                  {conn.label && <text x={midX} y={midY - 5} textAnchor="middle" fontSize="12" fill="black" pointerEvents="none">{conn.label}</text>}
                </g>
              );
            })}
            {tempLine && <line x1={tempLine.x1} y1={tempLine.y1} x2={tempLine.x2} y2={tempLine.y2} stroke="gray" strokeWidth="2" strokeDasharray="5,5" markerEnd="url(#arrowhead)" />}
          </svg>

          <div style={{ transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`, transformOrigin: "top left" }}>
            {nodes.map((n) => (
              <NodeCard key={n.id} node={n} isSelected={selectedNodeIds.includes(n.id)} onDelete={() => removeNode(n.id)} onMouseDown={(e) => startDrag(e, n.id)} onStartConnection={() => setLineStartNode(n.id)} onClick={(e) => handleNodeClick(e, n.id)} />
            ))}
          </div>
        </div>
      </div>

      {selectedNode && <PropertiesPanel node={selectedNode} connections={connections} onUpdateConnection={handleUpdateConnection} onUpdateNode={handleUpdateNode} />}
    </div>
  );
}
