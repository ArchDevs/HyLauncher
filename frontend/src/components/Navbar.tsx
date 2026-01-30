// src/components/Navbar.tsx
import React from "react";
import { pages } from "../config/pages";

interface NavbarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

function Navbar({ activeTab, onTabChange }: NavbarProps) {
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
        pointer-events-auto
        z-50
      "
    >
      {/* TOP ICONS */}
      <div className="flex flex-col items-center gap-6">
        {pages.map((page) => {
          const Icon = page.icon;
          const isActive = activeTab === page.id;
          return (
            <button
              key={page.id}
              onClick={() => {
                console.log("Navbar click:", page.id);
                onTabChange(page.id);
              }}
              style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
              className={`transition-all cursor-pointer pointer-events-auto ${
                isActive
                  ? "text-[#FFFFFF]/[0.90]"
                  : "text-[#FFFFFF]/[0.50] hover:text-[#FFFFFF]/[0.70]"
              }`}
              title={page.name}
            >
              <Icon size={18} />
            </button>
          );
        })}
      </div>
    </div>
  );
}

export default Navbar;
