// src/components/Navbar.tsx
import React, { useContext } from "react";
import { getPages } from "../config/pages";
import { useTranslation } from "../i18n";
import telegramIcon from "../assets/images/telegram.svg";
import discordIcon from "../assets/images/discord.svg";
import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";
import { OpenFolder } from "../../wailsjs/go/app/App";
import { Activity, Bolt, FolderOpen } from "lucide-react";
import { SettingsOverlayContext } from "../context/SettingsOverlayContext";

interface NavbarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  onDiagnosticsClick?: () => void;
  onSettingsClick?: () => void;
}

function Navbar({ activeTab, onTabChange, onDiagnosticsClick, onSettingsClick }: NavbarProps) {
  const { t } = useTranslation();
  const pages = getPages(t);
  const isSettingsOpen = useContext(SettingsOverlayContext);

  const openTelegram = () => {
    try {
      BrowserOpenURL("https://t.me/hylauncher");
    } catch {
      window.open("https://t.me/hylauncher", "_blank");
    }
  };

  const openDiscord = () => {
    try {
      BrowserOpenURL("https://dsc.gg/hylauncher");
    } catch {
      window.open("https://dsc.gg/hylauncher", "_blank");
    }
  };

  return (
    <div
      className="
        absolute left-[20px] top-1/2 -translate-y-1/2
        w-[48px] h-[324px]
        bg-[#090909]/[0.55]
        backdrop-blur-[12px]
        rounded-[14px]
        border border-[#7C7C7C]/[0.10]
        p-[16px]
        flex flex-col
        pointer-events-auto
        z-[110]
      "
    >
      {/* TOP ICONS */}
      <div className="flex flex-col items-center gap-[16px]">
        {pages.map((page) => {
          const Icon = page.icon;
          const isActive = !isSettingsOpen && activeTab === page.id;
          const isDisabled = page.id === "mods";
          return (
            <button
              key={page.id}
              onClick={() => {
                if (isDisabled) return;
                console.log("Navbar click:", page.id);
                onTabChange(page.id);
              }}
              disabled={isDisabled}
              style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
              className={`transition-all pointer-events-auto ${
                isDisabled
                  ? "cursor-not-allowed text-neutral-600 opacity-30"
                  : "cursor-pointer text-white " + (isActive ? "opacity-90" : "opacity-50 hover:opacity-70")
              }`}
              title={isDisabled ? "Coming soon" : page.name}
            >
              <Icon size={18} />
            </button>
          );
        })} 
        {/* Divider */}
        <div
          className="w-[48px] h-[1px] bg-[#7C7C7C]/[0.10]"
          style={{ marginLeft: 0, marginRight: 0 }}
        />
        {/* Telegram icon */}
        <button
          type="button"
          onClick={openTelegram}
          style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
          className="transition-all w-[18px] h-[18px] cursor-pointer pointer-events-auto opacity-60 hover:opacity-90"
          title="Telegram"
        >
          <img src={telegramIcon} alt="Telegram" />
        </button>
        {/* Discord icon */}
        <button
          type="button"
          onClick={openDiscord}
          style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
          className="transition-all w-[18px] h-[18px] cursor-pointer pointer-events-auto opacity-60 hover:opacity-90"
          title="Discord"
        >
          <img src={discordIcon} alt="Discord" />
        </button>
        <div
          className="w-[48px] h-[1px] bg-[#D9D9D9]/[0.10]"
          style={{ marginLeft: 0, marginRight: 0 }}
        />
        {/* Diagnostics icon */}
        <button
          type="button"
          onClick={onDiagnosticsClick}
          disabled={!onDiagnosticsClick}
          style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
          className={`transition-all cursor-not-allowed pointer-events-auto text-neutral-700 ${
            onDiagnosticsClick ? "opacity-30" : "opacity-60 hover:opacity-90"
          }`}
          title="Диагностика"
        >
          <Activity size={18} />
        </button>
        <button
          type="button"
          onClick={OpenFolder}
          style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
          className="transition-all cursor-pointer pointer-events-auto text-white opacity-60 hover:opacity-90"
          title="Папка игры"
        >
          <FolderOpen size={18} />
        </button>
        <button
          type="button"
          onClick={onSettingsClick}
          style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
          className={`transition-all cursor-pointer pointer-events-auto text-white ${
            isSettingsOpen ? "opacity-90" : "opacity-50 hover:opacity-70"
          }`}
          title="Настройки"
        >
          <Bolt size={18} />
        </button>
      </div>
    </div>
  );
}

export default Navbar;
