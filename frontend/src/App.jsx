import React from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";

import WorkflowBuilder from "./pages/WorkflowBuilder";
import Dashboard from "./pages/Dashboard";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/builder" element={<WorkflowBuilder />} />
      </Routes>
    </BrowserRouter>
  );
}
