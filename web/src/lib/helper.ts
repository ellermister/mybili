export const image = (path: string) => {
    return `/storage/images/${path}`
}

export const formatTimestamp = (timestamp: number, format: string) => {
    const date = new Date(timestamp * 1000);

    const map = {
        'yyyy': date.getFullYear(),
        'mm': String(date.getMonth() + 1).padStart(2, '0'),
        'dd': String(date.getDate()).padStart(2, '0'),
        'hh': String(date.getHours()).padStart(2, '0'),
        'ii': String(date.getMinutes()).padStart(2, '0'),
        'ss': String(date.getSeconds()).padStart(2, '0'),
    };

    return format.replace(/yyyy|mm|dd|hh|ii|ss/g, matched => map[matched]);
}



export const getLocale = () => {
    const locale = localStorage.getItem('locale');
    if (locale && ['zh-CN', 'en-US'].includes(locale)) {
        return locale;
    }
    return navigator.language;
}