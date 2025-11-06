import React from "react";
import { Link } from "react-router-dom";

export default function Dashboard() {
  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-6xl mx-auto">
        <header className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-bold">AutoFlow.AI</h1>
          <nav>
            <Link to="/builder" className="px-4 py-2 bg-indigo-600 text-white rounded-lg">Open Builder</Link>
          </nav>
        </header>

        <section className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="col-span-2 bg-white p-6 rounded-2xl shadow">
            <h2 className="text-lg font-semibold mb-2">Welcome</h2>
            <p className="text-sm text-gray-600">This is the Dashboard. Click <span className="font-semibold">Open Builder</span> to start building workflows.</p>
          </div>

          <aside className="bg-white p-6 rounded-2xl shadow">
            <h3 className="text-sm font-medium text-gray-700">Quick Links</h3>
            <ul className="mt-3 space-y-2 text-sm">
              <li>Recent workflows</li>
              <li>Templates</li>
              <li>Settings</li>
            </ul>
          </aside>
        </section>
      </div>
    </div>
  );
}
