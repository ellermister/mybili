export const getSystemInfo = () => {
    return fetch('/api/system/info').then((res) => res.json());
};

export const getMediaUsage = () => {
    return fetch('/api/system/media-usage').then((res) => res.json());
};

export const migrateDatabase = () => {
    return fetch('/api/system/migrate', { method: 'POST' }).then(async (res) => {
        const data = await res.json();
        if (!res.ok) {
            throw new Error(data.message ?? 'migrate failed');
        }
        return data as { message: string };
    });
};