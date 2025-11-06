import React from "react";

export default function Topbar() {
  return (
    <header className="w-full bg-white shadow-sm">
      <div className="max-w-[1400px] mx-auto px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-indigo-600 rounded-full flex items-center justify-center text-white font-bold">AF</div>
          <h2 className="text-lg font-semibold">AutoFlow.AI</h2>
        </div>

        <div className="flex items-center gap-3">
          <input placeholder="Search workflows..." className="border rounded-lg px-3 py-2 text-sm" />
          <button className="px-3 py-2 border rounded-lg text-sm">Account</button>
        </div>
      </div>
    </header>
  );
}
