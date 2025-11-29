import React, { useState } from 'react';
import { Wheel } from 'react-custom-roulette';
import type { Item } from '../App';

interface WheelComponentProps {
    items: Item[];
    isDarkMode: boolean;
}

const WheelComponent: React.FC<WheelComponentProps> = ({ items, isDarkMode }) => {
    const [mustSpin, setMustSpin] = useState(false);
    const [prizeNumber, setPrizeNumber] = useState(0);

    const handleSpinClick = () => {
        if (!mustSpin) {
            const newPrizeNumber = Math.floor(Math.random() * items.length);
            setPrizeNumber(newPrizeNumber);
            setMustSpin(true);
        }
    };

    const lightColors = {
        background: ['#f5f5f7', '#ffffff'],
        text: '#1d1d1f',
        border: '#d2d2d7',
    };

    const darkColors = {
        background: ['#1d1d1f', '#000000'],
        text: '#f5f5f7',
        border: '#424245',
    };

    const colors = isDarkMode ? darkColors : lightColors;

    return (
        <div className="relative rounded-xl w-full max-w-sm md:max-w-md lg:max-w-lg h-full bg-fill flex items-center justify-center overflow-hidden transition-colors duration-300 p-4">
            <div className="relative z-20 flex flex-col items-center justify-center gap-10 px-6">
                <div className="rounded-full border-8 border-border shadow-lg transition-transform duration-300 hover:scale-105">
                    <Wheel
                        mustStartSpinning={mustSpin}
                        prizeNumber={prizeNumber}
                        data={items}
                        onStopSpinning={() => setMustSpin(false)}
                        backgroundColors={colors.background}
                        textColors={[colors.text]}
                        disableInitialAnimation={true}
                        outerBorderColor={colors.border}
                        outerBorderWidth={14}
                        innerBorderColor={colors.border}
                        innerBorderWidth={10}
                        radiusLineColor={colors.border}
                        fontSize={20}
                        spinDuration={0.5}
                    />
                </div>
                <button
                    onClick={handleSpinClick}
                    className="px-12 py-3 rounded-full bg-accent text-white font-semibold text-lg shadow-lg hover:opacity-90 transition-opacity active:scale-95"
                >
                    Spin
                </button>
            </div>
        </div>
    );
};

export default WheelComponent;
