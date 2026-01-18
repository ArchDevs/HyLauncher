import { useState, useEffect } from 'react';
// @ts-ignore
import { WindowIsMaximised, EventsOn } from '../../wailsjs/runtime/runtime';

export function useWindowState() {
    const [isMaximised, setIsMaximised] = useState(false);

    useEffect(() => {
        let timeoutId: any;

        const updateState = () => {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => {
                WindowIsMaximised().then(setIsMaximised);
            }, 100);
        };

        // Initial check
        updateState();

        // Check on resize (handles snap, double-click titlebar, maximizing)
        window.addEventListener('resize', updateState);

        // Also listen to Wails events directly for faster feedback
        const unmountMaximise = EventsOn("window:maximise", () => setIsMaximised(true));
        const unmountUnmaximise = EventsOn("window:unmaximise", () => setIsMaximised(false));

        return () => {
            window.removeEventListener('resize', updateState);
            clearTimeout(timeoutId);
            unmountMaximise();
            unmountUnmaximise();
        };
    }, []);

    return isMaximised;
}
