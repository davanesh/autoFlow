import React from "react";
import Topbar from "../components/Topbar";
import Sidebar from "../components/Siderbr";
import CanvasArea from "../components/CanvasArea";

export default function WorkflowBuilder() {
  return (
    <div className="min-h-screen bg-gray-100">
      <Topbar />
      <div className="max-w-[1400px] mx-auto p-4 grid grid-cols-12 gap-4">
        <aside className="col-span-3">
          <Sidebar />
        </aside>

        <main className="col-span-9">
          <CanvasArea />
        </main>
      </div>
    </div>
  );
}
