import React from "react";
import BackgroundImage from "./components/BackgroundImage";
import Titlebar from "./components/Titlebar";
import Navbar from "./components/Navbar";
import HomePage from "./pages/Home";
import { BrowserOpenURL } from "../wailsjs/runtime/runtime";

const App: React.FC = () => {
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

      <Navbar />
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
      <HomePage />
    </div>
  );
};

export default App;
