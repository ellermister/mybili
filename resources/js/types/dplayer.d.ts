declare module 'dplayer' {
    interface DPlayerOptions {
        container: HTMLElement;
        video: {
            url: string;
            type: string;
        };
        danmaku: {
            id: string;
            api: string;
        };
    }

    interface DPlayerInstance {
        switchVideo: (video: any, danmaku: any) => void;
        play: () => void;
        pause: () => void;
        on: (event: string, callback: () => void) => void;
        off: (event: string, callback: () => void) => void;
        danmaku: {
            draw: (options: any) => void;
        };
    }

    class DPlayer {
        constructor(options: DPlayerOptions);
    }

    export default DPlayer;
} 