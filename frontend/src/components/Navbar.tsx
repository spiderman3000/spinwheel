import React from 'react';

interface NavbarProps {
    isDarkMode: boolean;
    toggleTheme: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ isDarkMode, toggleTheme }) => {
    return (
        <div className="w-full z-50 bg-fill border-b border-border transition-colors">
            <div className="max-w-7xl mx-auto px-6 sm:px-10 py-4 flex items-center justify-between">
                <span className="text-xl font-semibold text-text-primary tracking-tight select-none">
                    Spinwheel
                </span>
                <button
                    onClick={toggleTheme}
                    className="p-2 rounded-full text-text-primary text-lg"
                >
                    {isDarkMode ? '🌞' : '🌜'}
                </button>
            </div>
        </div>
    );
};
