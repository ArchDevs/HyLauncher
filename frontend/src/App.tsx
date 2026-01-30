import React, { useState, useEffect } from "react";
import BackgroundImage from "./components/BackgroundImage";
import Titlebar from "./components/Titlebar";
import Navbar from "./components/Navbar";
import { BrowserOpenURL } from "../wailsjs/runtime/runtime";
import { AnimatePresence, motion } from "framer-motion";
import { getDefaultPage, getPageById } from "./config/pages";

const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState(getDefaultPage().id);
  
  // Debug: Log when activeTab changes
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

  return (
    <div className="relative w-screen h-screen max-w-[1280px] max-h-[720px] bg-[#090909] text-white overflow-hidden font-sans select-none rounded-[14px] border border-white/5 mx-auto">
      {/* Background doesn't intercept clicks */}
      <div className="absolute inset-0 pointer-events-none">
        <BackgroundImage />
      </div>

      <Navbar activeTab={activeTab} onTabChange={setActiveTab} />
      <Titlebar />

      <button
        type="button"
        onClick={openExternal}
        style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
        className="
          absolute left-[20px] top-[60px]
          w-[48px] h-[48px]
          bg-[#090909]/55 backdrop-blur-[12px]
          rounded-[14px] border border-[#7C7C7C]/[0.10]
          tracking-[-3%] cursor-pointer
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

      {/* Page Content */}
      <AnimatePresence mode="wait">
        {(() => {
          const page = getPageById(activeTab);
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
              transition={{ duration: 0.2 }}
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
