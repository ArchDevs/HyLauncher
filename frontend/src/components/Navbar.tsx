// src/components/Navbar.tsx
import React from "react";
import { getPages } from "../config/pages";
import { useTranslation } from "../i18n";
import telegramIcon from "../assets/images/telegram.png";

interface NavbarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

function Navbar({ activeTab, onTabChange }: NavbarProps) {
  const { t } = useTranslation();
  const pages = getPages(t);

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
      <div className="flex flex-col items-center gap-[16px]">
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
              className={`transition-all cursor-pointer pointer-events-auto text-white ${
                isActive ? "opacity-90" : "opacity-50 hover:opacity-70"
              }`}
              title={page.name}
            >
              <Icon size={18} />
            </button>
          );
        })}
        {/* Divider */}
        <div
          className="w-[48px] h-[1px] bg-[#D9D9D9]/[0.10]"
          style={{ marginLeft: 0, marginRight: 0 }}
        />
        {/* Telegram icon */}
        <a
          href="https://t.me/hylauncher"
          target="_blank"
          rel="noopener noreferrer"
          style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
          className="transition-all cursor-pointer pointer-events-auto opacity-60 hover:opacity-90"
          title="HyLauncher Telegram"
        >
          <img src={telegramIcon} alt="Telegram"/>
        </a>
      </div>
    </div>
  );
}

export default Navbar;
