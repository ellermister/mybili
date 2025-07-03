export const getSettings = () => {
    return fetch('/api/settings').then((res) => res.json());
};

export const saveSettings = (settings: any) => {
    return fetch('/api/settings', {
        method: 'POST',
        body: JSON.stringify(settings),
    });
};

export const testTelegramConnection = (botToken: string, chatId: string) => {
    return fetch('/api/settings/test-telegram', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            telegram_bot_token: botToken,
            telegram_chat_id: chatId,
        }),
    }).then((res) => res.json());
};
