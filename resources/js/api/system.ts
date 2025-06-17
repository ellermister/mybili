export const getSystemInfo = () => {
    return fetch('/api/system/info').then((res) => res.json());
};