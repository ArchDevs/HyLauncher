import { Minus, X } from "lucide-react";
// @ts-ignore
import { Quit, WindowMinimise } from "../../wailsjs/runtime/runtime";

function Titlebar() {
  return (
    <div className="absolute top-0 left-0 w-full h-[28px] z-50 flex justify-end items-center pl-4 pr-3 bg-[#090909]/[0.65] backdrop-blur-xl border-b border-white/5 [--wails-draggable:drag] font-[Mazzard] text-[12px] text-[#CCD9E0]/[0.65]">
      <div className="absolute left-0 h-full flex items-center pl-[12px] no-drag [--wails-draggable:no-drag]">
        <span className="font-[Mazzard] text-[12px] text-[#CCD9E0]/[0.65] leading-none">
          HyLauncher | Online: 0
        </span>
        <span
          aria-hidden
          className="ml-1.5 w-1.5 h-1.5 rounded-full bg-[#5FE53E]/[0.90]"
        />
      </div>

      <div className="flex gap-4 no-drag [--wails-draggable:no-drag]">
        <button
          onClick={() => WindowMinimise()}
          className="cursor-pointer gray-500 hover:text-white transition-colors"
        >
          <Minus size={16} />
        </button>
        <button
          onClick={() => Quit()}
          className="cursor-pointer gray-500 hover:text-red-500 transition-colors"
        >
          <X size={16} />
        </button>
      </div>
    </div>
  );
}

export default Titlebar;
