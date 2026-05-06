import { useEffect, useState } from 'react';

export default function ThemeToggle() {
    // We set the default state to dark mode
    const [isDark, setIsDark] = useState(true);

    // This runs once when the page loads to check if they saved a preference
    useEffect(() => {
        const savedTheme = localStorage.getItem('theme');
        if (savedTheme === 'light') {
            setIsDark(false);
            document.documentElement.classList.remove('dark');
        } else {
            document.documentElement.classList.add('dark');
            localStorage.setItem('theme', 'dark');
        }
    }, []);

    // This function runs when the button is clicked
    const toggleTheme = () => {
        if (isDark) {
            document.documentElement.classList.remove('dark');
            localStorage.setItem('theme', 'light');
            setIsDark(false);
        } else {
            document.documentElement.classList.add('dark');
            localStorage.setItem('theme', 'dark');
            setIsDark(true);
        }
    };

    return (
        <button 
            onClick={toggleTheme} 
            className="p-2 flex items-center justify-center rounded-lg border border-[#342720] hover:bg-[#342720] transition-colors bubble-pop"
            title="Toggle Light/Dark Mode"
        >
            <span className="material-symbols-outlined text-[#e7b4ff]">
                {isDark ? 'light_mode' : 'dark_mode'}
            </span>
        </button>
    );
}