export const getSettings = () => {
    return fetch('/api/settings').then((res) => res.json());
};

export const saveSettings = async (settings: any) => {
    const res = await fetch('/api/settings', {
        method: 'POST',
        body: JSON.stringify(settings),
    })
    if (!res.ok) {
        const data = await res.json();
        console.error(data);
        throw new Error(data.message ?? 'Failed to save settings');
    }
    return await res.json();
};

export const testTelegramConnection = (botToken: string, chatId: string, botUrl?: string) => {
    return fetch('/api/settings/test-telegram', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            telegram_bot_token: botToken,
            telegram_chat_id: chatId,
            telegram_bot_api_url: botUrl,
        }),
    }).then((res) => res.json());
};
