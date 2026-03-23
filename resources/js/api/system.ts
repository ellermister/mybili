export const getSystemInfo = () => {
    return fetch('/api/system/info').then((res) => res.json());
};

export const getMediaUsage = () => {
    return fetch('/api/system/media-usage').then((res) => res.json());
};