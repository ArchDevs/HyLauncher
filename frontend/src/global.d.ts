export { };

declare global {
    interface Window {
        go: any;
        [key: string]: any; // This allows window['go'], window['runtime'], etc.
    }
}
