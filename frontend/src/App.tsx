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

  useEffect(() => {
    console.log("Active tab changed to:", activeTab);
  }, [activeTab]);

  useEffect(() => {
    setShowSettingsOverlay(false);
  }, [activeTab]);

  useEffect(() => {
    if (!showSettingsOverlay) setOverlayEnterReady(false);
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
                <span className="text-[20px] font-[Unbounded] font-[600] uppercase tracking-wide text-white/90">
                  SETTINGS
                </span>
                <span className="w-[1px] h-[20px] bg-[#7C7C7C]/[0.10]" aria-hidden />
                <span className="text-[14px] text-white/40 font-[Mazzard]">
                  HyLauncher v{launcherVersion}
                </span>
              </div>
              {/* Горизонтальная разделительная полоска под шапкой */}
              <div
                className="absolute left-0 right-0 top-[80px] z-10 h-[1px] bg-[#7C7C7C]/[0.10]"
                aria-hidden
              />
              {/* Вертикальная разделительная линия (сайдбар / контент) */}
              <div
                className="absolute left-[176px] top-[80px] bottom-0 z-10 w-[1px] bg-[#7C7C7C]/[0.10]"
                aria-hidden
              />
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
