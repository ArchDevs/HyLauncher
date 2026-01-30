// src/components/Navbar.tsx
import React from "react";
import { Gamepad2, Globe } from "lucide-react";

function Navbar() {
  return (
    <div
      className="
        absolute left-[20px] top-1/2 -translate-y-1/2
        w-[48px] h-[306px]
        bg-[#090909]/[0.55]
        backdrop-blur-[12px]
        rounded-[14px]
        border border-[#7C7C7C]/[0.10]
        p-[16px]
        flex flex-col
      "
    >
      {/* TOP ICONS */}
      <div className="flex flex-col items-center gap-6">
        <button className="text-[#FFFFFF]/[0.90]">
          <Gamepad2 size={18} />
        </button>

        <button className="text-[#FFFFFF]/[0.50]">
          <Globe size={18} />
        </button>
      </div>
    </div>
  );
}

export default Navbar;
