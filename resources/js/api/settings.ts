export const getSettings = () => {
    return fetch('/api/settings').then((res) => res.json());
};

export const saveSettings = (settings: any) => {
    return fetch('/api/settings', {
        method: 'POST',
        body: JSON.stringify(settings),
    });
};
