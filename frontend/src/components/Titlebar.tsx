import React from "react";
import { Minus, Square, Copy, X } from "lucide-react";
// @ts-ignore
import { Quit, WindowMinimise, WindowToggleMaximise } from '../../wailsjs/runtime/runtime';

interface TitlebarProps {
    isMaximised: boolean;
}

function Titlebar({ isMaximised }: TitlebarProps) {

    return (
        <div
            style={{ "--wails-draggable": "drag" } as any}
            // Colors: #090909BF (75% Opacity)
            className="absolute top-0 left-0 w-full h-[28px] z-50 flex justify-end items-center px-4 bg-[#090909]/[0.75] backdrop-blur-xl border-b border-white/5"
        >
            <div className="flex gap-4 no-drag" style={{ "--wails-draggable": "no-drag" } as any}>
                <button
                    onClick={() => WindowMinimise()}
                    className="cursor-pointer text-gray-500 hover:text-white transition-colors"
                >
                    <Minus size={16} />
                </button>
                <button
                    onClick={() => WindowToggleMaximise()}
                    className="cursor-pointer text-gray-500 hover:text-white transition-colors"
                >
                    {isMaximised ? (
                        <Copy size={12} className="rotate-90" />
                    ) : (
                        <Square size={14} />
                    )}
                </button>
                <button
                    onClick={() => Quit()}
                    className="cursor-pointer text-gray-500 hover:text-red-500 transition-colors"
                >
                    <X size={16} />
                </button>
            </div>
        </div>
    );
}

export default Titlebar;