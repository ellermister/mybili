// Author: N6005 Hermes (Hermes Agent)
// Purpose: Persist and apply the light/dark theme choice across pages and reloads.
import { ref, watch } from 'vue';

const STORAGE_KEY = 'theme';

const isDark = ref(document.documentElement.classList.contains('dark'));

watch(isDark, (value) => {
    document.documentElement.classList.toggle('dark', value);
    localStorage.setItem(STORAGE_KEY, value ? 'dark' : 'light');
});

export function useTheme() {
    const toggleTheme = () => {
        isDark.value = !isDark.value;
    };

    return { isDark, toggleTheme };
}
