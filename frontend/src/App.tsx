import React, { useState, useEffect } from "react";
import Titlebar from "./components/Titlebar";
import Navbar from "./components/Navbar";
import { BrowserOpenURL } from "../wailsjs/runtime/runtime";
import { AnimatePresence, motion, cubicBezier } from "framer-motion";
import { getDefaultPage, getPageById } from "./config/pages";

const bgTransition = {
  duration: 0.45, // macOS-speed
  ease: cubicBezier(0.16, 1, 0.3, 1),
};

const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState(getDefaultPage().id);

  useEffect(() => {
    console.log("Active tab changed to:", activeTab);
  }, [activeTab]);

  const openExternal = () => {
    try {
      BrowserOpenURL("https://github.com/ArchDevs/HyLauncher");
    } catch {
      window.open("https://github.com/ArchDevs/HyLauncher/", "_blank");
    }
  };

  const page = getPageById(activeTab);
  const Background = page?.background;

  return (
    <div className="relative w-screen h-screen max-w-[1280px] max-h-[720px] bg-[#090909] text-white overflow-hidden font-sans select-none rounded-[14px] border border-white/5 mx-auto">
      {/* BACKGROUND */}
      <div className="absolute inset-0 pointer-events-none">
        <AnimatePresence mode="wait">
          <motion.div
            key={activeTab}
            className="absolute inset-0"
            initial={{
              opacity: 0,
              scale: 1.035,
              filter: "blur(14px)",
            }}
            animate={{
              opacity: 1,
              scale: 1,
              filter: "blur(0px)",
            }}
            exit={{
              opacity: 0,
              scale: 1.02,
              filter: "blur(10px)",
            }}
            transition={bgTransition}
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

      <Navbar activeTab={activeTab} onTabChange={setActiveTab} />
      <Titlebar />

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
          src="src/assets/images/logo.png"
          alt="Logo"
          className="w-[40px] h-[40px] pointer-events-none"
          draggable={false}
        />
      </button>

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
  );
};

export default App;
