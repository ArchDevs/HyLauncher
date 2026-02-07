import React, { useState, useEffect, useLayoutEffect } from "react";
import Titlebar from "./components/Titlebar";
import Navbar from "./components/Navbar";
import { LanguageSwitcher } from "./components/LanguageSwitcher";
import { BrowserOpenURL } from "../wailsjs/runtime/runtime";
import { AnimatePresence, motion, cubicBezier } from "framer-motion";
import { getDefaultPage, getPageById } from "./config/pages";
import logoImage from "./assets/images/logo.png";
import bgSettingsImage from "./assets/images/bg-settings.png";
import { useTranslation } from "./i18n";
import { useLauncher } from "./hooks/useLauncher";
import { X, HardDrive, Shield, Languages, FolderOpen, FolderSearch, Trash2 } from "lucide-react";

const bgTransition = {
  duration: 0.45, // macOS-speed
  ease: cubicBezier(0.16, 1, 0.3, 1),
};

import { SettingsOverlayContext } from "./context/SettingsOverlayContext";

const App: React.FC = () => {
  const { t } = useTranslation();
  const { launcherVersion } = useLauncher();
  const [activeTab, setActiveTab] = useState(getDefaultPage(t).id);
  const [showSettingsOverlay, setShowSettingsOverlay] = useState(false);
  const [overlayEnterReady, setOverlayEnterReady] = useState(false);
  type SettingsSection = "storage" | "privacy" | "language";
  const [settingsSection, setSettingsSection] = useState<SettingsSection>("storage");

  useEffect(() => {
    console.log("Active tab changed to:", activeTab);
  }, [activeTab]);

  useEffect(() => {
    setShowSettingsOverlay(false);
  }, [activeTab]);

  useEffect(() => {
    if (!showSettingsOverlay) {
      setOverlayEnterReady(false);
    } else {
      setSettingsSection("storage");
    }
  }, [showSettingsOverlay]);

  useLayoutEffect(() => {
    if (!showSettingsOverlay) return;
    setOverlayEnterReady(false);
    const raf = requestAnimationFrame(() => setOverlayEnterReady(true));
    return () => cancelAnimationFrame(raf);
  }, [showSettingsOverlay]);

  // Закрытие настроек по ESC (глобальный слушатель, не зависит от фокуса)
  useEffect(() => {
    if (!showSettingsOverlay) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setShowSettingsOverlay(false);
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [showSettingsOverlay]);

  const openExternal = () => {
    try {
      BrowserOpenURL("https://github.com/ArchDevs/HyLauncher");
    } catch {
      window.open("https://github.com/ArchDevs/HyLauncher/", "_blank");
    }
  };

  const page = getPageById(activeTab, t);
  const Background = page?.background;

  return (
    <SettingsOverlayContext.Provider value={showSettingsOverlay}>
    <div className="relative w-screen h-screen max-w-[1280px] max-h-[720px] bg-[#090909] text-white overflow-hidden font-sans select-none rounded-[14px] border border-white/5 mx-auto">
      {/* BACKGROUND */}
      <div className="absolute inset-0 pointer-events-none">
        <AnimatePresence mode="wait">
          <motion.div
            key={activeTab}
            className="absolute inset-0"
            style={{
              willChange: "opacity, transform, filter",
              transform: "translateZ(0)",
              backfaceVisibility: "hidden",
              perspective: "1000px",
            }}
            initial={{
              opacity: 0,
              scale: 1.02,
            }}
            animate={{
              opacity: 1,
              scale: 1,
            }}
            exit={{
              opacity: 0,
              scale: 1.01,
            }}
            transition={{
              ...bgTransition,
              filter: { duration: 0 },
            }}
          >
            {/* Page background */}
            {Background ? <Background /> : null}

            {/* macOS-style overlays */}
            <div className="absolute inset-0 bg-gradient-to-b from-black/20 via-transparent to-black/35" />
            <div className="absolute inset-0 [box-shadow:inset_0_0_120px_rgba(0,0,0,0.35)]" />
            <div className="absolute inset-0 opacity-[0.06] mix-blend-overlay noise-layer" />
          </motion.div>
        </AnimatePresence>
      </div>

      <Navbar
        activeTab={activeTab}
        onTabChange={setActiveTab}
        onSettingsClick={() => setShowSettingsOverlay(true)}
      />
      <Titlebar />
      
      {/* Launcher version */}
      <div className="absolute right-[16px] bottom-[16px] text-[#FFFFFF]/[0.25] text-[14px] font-[Mazzard]">
        v{launcherVersion}
      </div>

      {/* GitHub button */}
      <button
        type="button"
        onClick={openExternal}
        style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
        className="
          absolute left-[20px] top-[60px]
          w-[48px] h-[48px]
          bg-[#090909]/55 backdrop-blur-[12px]
          rounded-[14px] border border-[#7C7C7C]/[0.10]
          cursor-pointer
          flex items-center justify-center
          active:scale-95
          transition-all duration-150
          z-[9999] pointer-events-auto
        "
      >
        <img
          src={logoImage}
          alt="Logo"
          className="w-[40px] h-[40px] pointer-events-none"
          draggable={false}
        />
      </button>

      {/* Language Switcher */}
      <div className="absolute left-[20px] top-[116px] z-[9999] pointer-events-none">
        <LanguageSwitcher />
      </div>

      {/* Settings overlay: сначала блюр (отдельная анимация), потом окно */}
      <AnimatePresence mode="wait">
        {showSettingsOverlay && (
          <motion.div
            key="settings-overlay"
            initial={{ opacity: 1 }}
            animate={{ opacity: 1 }}
            exit={{
              opacity: 0,
              scale: 0.98,
              transition: { duration: 0.28, ease: [0.16, 1, 0.3, 1] },
            }}
            className="absolute inset-0 z-[100] flex items-center justify-center origin-center"
            tabIndex={0}
            onKeyDown={(e) => e.key === "Escape" && setShowSettingsOverlay(false)}
          >
            {/* 1) Появление блюра */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{
                opacity: overlayEnterReady ? 1 : 0,
                transition: { duration: 0.4, ease: [0.22, 1, 0.36, 1] },
              }}
              className="absolute inset-0 bg-black/25 backdrop-blur-[14px] cursor-default"
              onClick={() => setShowSettingsOverlay(false)}
              aria-hidden
            />
            {/* 2) Окно */}
            <motion.div
              initial={{ opacity: 0, scale: 0.97 }}
              animate={{
                opacity: overlayEnterReady ? 1 : 0,
                scale: overlayEnterReady ? 1 : 0.97,
                transition: {
                  duration: 0.22,
                  delay: 0,
                  ease: [0.16, 1, 0.3, 1],
                },
              }}
              className="relative z-10 w-[900px] h-[500px] rounded-[14px] border border-[#7C7C7C]/[0.10] bg-cover bg-center bg-no-repeat overflow-hidden"
              style={{ backgroundImage: `url(${bgSettingsImage})` }}
              onClick={(e) => e.stopPropagation()}
              role="dialog"
              aria-label="Настройки"
            >
              {/* Лёгкий затемняющий слой поверх фона для читаемости контента */}
              <div className="absolute inset-0 bg-[#090909]/[0.75] rounded-[14px]" aria-hidden />
              {/* Заголовок SETTINGS | HyLauncher v... */}
              <div className="absolute left-[30px] top-[30px] z-10 flex items-center gap-[12px]">
                <span className="text-[20px] font-[Unbounded] font-[500] uppercase tracking-wide text-white/90">
                  SETTINGS
                </span>
                <span className="w-[1px] h-[20px] bg-[#7C7C7C]/[0.10]" aria-hidden />
                <span className="text-[14px] text-white/25 font-[Mazzard]">
                  HyLauncher v{launcherVersion}
                </span>
              </div>
              {/* Горизонтальная разделительная полоска под шапкой */}
              <div
                className="absolute left-0 right-0 top-[80px] z-10 h-[1px] bg-[#7C7C7C]/[0.10]"
                aria-hidden
              />
              {/* Вертикальная линия начинается с 81px, чтобы не накладываться на горизонтальную */}
              <div
                className="absolute left-[176px] top-[81px] bottom-0 z-10 w-[1px] bg-[#7C7C7C]/[0.10]"
                aria-hidden
              />
              {/* Сайдбар: разделы настроек */}
              <div className="absolute left-[30px] top-[111px] z-10 flex flex-col gap-[12px]">
                <button
                  type="button"
                  onClick={() => setSettingsSection("storage")}
                  className={`flex items-center gap-2 px-2 py-0 rounded text-left transition-opacity cursor-pointer text-white ${
                    settingsSection === "storage" ? "opacity-90" : "opacity-50 active:opacity-90"
                  }`}
                  aria-label="Storage"
                >
                  <HardDrive size={18} strokeWidth={2} />
                  <span className="text-[16px] font-[Mazzard]">Storage</span>
                </button>
                <button
                  type="button"
                  onClick={() => setSettingsSection("privacy")}
                  className={`flex items-center gap-2 px-2 py-0 rounded text-left transition-opacity cursor-pointer text-white ${
                    settingsSection === "privacy" ? "opacity-90" : "opacity-50 active:opacity-90"
                  }`}
                  aria-label="Privacy"
                >
                  <Shield size={18} strokeWidth={2} />
                  <span className="text-[16px] font-[Mazzard]">Privacy</span>
                </button>
                <button
                  type="button"
                  onClick={() => setSettingsSection("language")}
                  className={`flex items-center gap-2 px-2 py-0 rounded text-left transition-opacity cursor-pointer text-white ${
                    settingsSection === "language" ? "opacity-90" : "opacity-50 active:opacity-90"
                  }`}
                  aria-label="Language"
                >
                  <Languages size={18} strokeWidth={2} />
                  <span className="text-[16px] font-[Mazzard]">Language</span>
                </button>
              </div>
              {/* Контент выбранного раздела (справа от сайдбара) */}
              <div className="absolute left-[206px] right-[30px] top-[111px] bottom-[30px] z-10 overflow-y-auto">
                {settingsSection === "storage" && (
                  <div className="flex flex-col gap-[24px] text-white/90 font-[Mazzard]">
                    <section>
                      <h3 className="text-[16px] font-[Unbounded] font-[500] text-white mb-[6px]">Game directory</h3>
                      <p className="text-[14px] font-[Mazzard] text-white/50 mb-[6px]">
                        The directory where the game stores all of its files. Changes will be applied after restarting the launcher.
                      </p>
                      <div className="relative w-full">
                        <input
                          type="text"
                          readOnly
                          className="w-full h-[46px] pl-4 pr-10 rounded-[14px] bg-[#090909]/[0.55] border border-[#7C7C7C]/[0.10] text-[14px] text-[#CCD9E0]/[0.9] font-[Mazzard]"
                          value="C:\Users\Admin RoNi\AppData\Local\HyLauncher"
                        />
                        <button
                          type="button"
                          className="absolute right-[16px] top-1/2 -translate-y-1/2 flex items-center justify-center w-8 h-8 text-[#CCD9E0]/[0.9] hover:opacity-80 transition-opacity"
                          aria-label="Browse"
                        >
                          <FolderSearch size={18} />
                        </button>
                      </div>
                    </section>
                    <section>
                      <h3 className="text-[16px] font-[Unbounded] font-[500] text-white mb-[6px]">Logs</h3>
                      <p className="text-[14px] font-[Mazzard] text-white/50 mb-[6px]">Browse or clean up your log files.</p>
                      <div className="flex items-center gap-[10px]">
                        <button
                          type="button"
                          className="flex items-center justify-center gap-[16px] w-[130px] h-[46px] rounded-[14px] bg-[#090909]/[0.55] border border-[#7C7C7C]/[0.10] font-[Mazzard] text-[#CCD9E0]/[0.9] text-[14px] hover:bg-[#090909]/70 transition-colors"
                        >
                          Open logs <FolderOpen size={16} />
                        </button>
                        <button
                          type="button"
                          className="flex items-center justify-center gap-[16px] w-[136px] h-[46px] rounded-[14px] bg-[#090909]/[0.55] border border-[#7C7C7C]/[0.10] font-[Mazzard] text-[#CCD9E0]/[0.9] text-[14px] hover:bg-[#090909]/70 transition-colors"
                        >
                          Delete logs <Trash2 size={16} />
                        </button>
                      </div>
                    </section>
                    <section>
                      <h3 className="text-[16px] font-[Unbounded] font-[500] text-white mb-[6px]">Clear Cache/Game</h3>
                      <p className="text-[14px] font-[Mazzard] text-white/50 mb-[6px]">
                        Clean up HyLauncher cache game files/full files game. (will temporarily increase launch time)
                      </p>
                      <div className="flex items-center gap-[10px]">
                        <button
                          type="button"
                          className="flex items-center justify-center gap-[16px] w-[154px] h-[46px] rounded-[14px] bg-[#090909]/[0.55] border border-[#7C7C7C]/[0.10] font-[Mazzard] text-[#CCD9E0]/[0.9] text-[14px] hover:bg-[#090909]/70 transition-colors"
                        >
                          Delete Cache <Trash2 size={16} />
                        </button>
                        <button
                          type="button"
                          className="flex items-center justify-center gap-[16px] w-[140px] h-[46px] rounded-[14px] bg-[#170000]/[0.55] border border-[#8F0000]/[0.10] font-[Mazzard] text-white text-[14px] hover:bg-[#170000]/70 transition-colors"
                        >
                          Delete Files <Trash2 size={16} />
                        </button>
                      </div>
                    </section>
                  </div>
                )}
                {settingsSection === "privacy" && (
                  <div className="text-white/70 font-[Mazzard] text-[14px]">
                    <h3 className="text-[16px] font-[Unbounded] font-[500] text-white mb-[6px]">Privacy</h3>
                    <p className="text-[14px] font-[Mazzard] text-white/50">Privacy settings will be here.</p>
                  </div>
                )}
                {settingsSection === "language" && (
                  <div className="text-white/70 font-[Mazzard] text-[14px]">
                    <h3 className="text-[16px] font-[Unbounded] font-[500] text-white mb-[6px]">Language</h3>
                    <p className="text-[14px] font-[Mazzard] text-white/50">Language settings will be here.</p>
                  </div>
                )}
              </div>
              {/* Крестик закрытия — сплошной цвет без opacity, чтобы не было накладки в пересечении линий */}
              <button
                type="button"
                onClick={() => setShowSettingsOverlay(false)}
                className="absolute right-[30px] top-[30px] z-10 flex items-center justify-center text-[#999999] transition-colors hover:text-white cursor-pointer"
                aria-label="Закрыть настройки"
              >
                <X size={18} strokeWidth={2} />
              </button>
              {/* HyLauncher <3 внизу слева */}
              <div
                className="absolute left-[30px] bottom-[30px] z-10 text-[12px] font-[Unbounded] text-white/25"
                aria-hidden
              >
                {"HyLauncher <3"}
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* PAGE CONTENT */}
      <AnimatePresence mode="wait">
        {(() => {
          if (!page) {
            console.error("Page not found for id:", activeTab);
            return null;
          }

          const PageComponent = page.component;

          return (
            <motion.div
              key={activeTab}
              style={{
                willChange: "opacity, transform",
                transform: "translateZ(0)",
                backfaceVisibility: "hidden",
              }}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              transition={{
                duration: 0.25,
                ease: cubicBezier(0.16, 1, 0.3, 1),
              }}
              className="h-full w-full"
            >
              <PageComponent />
            </motion.div>
          );
        })()}
      </AnimatePresence>
    </div>
    </SettingsOverlayContext.Provider>
  );
};

export default App;
