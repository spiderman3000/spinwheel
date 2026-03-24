import { useState, useRef, useEffect } from 'preact/hooks';
import type { Item } from '../App';

interface WheelProps {
    items: Item[];
    isDarkMode: boolean;
}

const Wheel = ({ items, isDarkMode }: WheelProps) => {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const [isSpinning, setIsSpinning] = useState(false);
    const [rotation, setRotation] = useState(0);
    
    const colors = isDarkMode 
        ? ['#1d1d1f', '#2c2c2e', '#3a3a3c', '#48484a'] 
        : ['#f5f5f7', '#e5e5ea', '#d1d1d6', '#c7c7cc'];
    
    const textColor = isDarkMode ? '#f5f5f7' : '#1d1d1f';
    const accentColor = '#0071e3';

    useEffect(() => {
        drawWheel();
    }, [items, rotation, isDarkMode]);

    const drawWheel = () => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        const size = canvas.width;
        const center = size / 2;
        const radius = size / 2 - 10;
        const sliceAngle = (2 * Math.PI) / items.length;

        ctx.clearRect(0, 0, size, size);

        // Draw slices
        items.forEach((item, i) => {
            const angle = rotation + i * sliceAngle;
            ctx.beginPath();
            ctx.moveTo(center, center);
            ctx.arc(center, center, radius, angle, angle + sliceAngle);
            ctx.closePath();
            
            ctx.fillStyle = colors[i % colors.length];
            ctx.fill();
            ctx.strokeStyle = isDarkMode ? '#424245' : '#d2d2d7';
            ctx.lineWidth = 2;
            ctx.stroke();

            // Draw text
            ctx.save();
            ctx.translate(center, center);
            ctx.rotate(angle + sliceAngle / 2);
            ctx.textAlign = 'right';
            ctx.fillStyle = textColor;
            ctx.font = 'bold 16px -apple-system, sans-serif';
            ctx.fillText(item.option, radius - 30, 6);
            ctx.restore();
        });

        // Draw outer border
        ctx.beginPath();
        ctx.arc(center, center, radius, 0, 2 * Math.PI);
        ctx.strokeStyle = isDarkMode ? '#424245' : '#d2d2d7';
        ctx.lineWidth = 8;
        ctx.stroke();

        // Draw center point
        ctx.beginPath();
        ctx.arc(center, center, 15, 0, 2 * Math.PI);
        ctx.fillStyle = accentColor;
        ctx.fill();
        ctx.strokeStyle = '#fff';
        ctx.lineWidth = 3;
        ctx.stroke();

        // Draw pointer (triangle)
        ctx.beginPath();
        ctx.moveTo(size - 5, center);
        ctx.lineTo(size - 25, center - 15);
        ctx.lineTo(size - 25, center + 15);
        ctx.closePath();
        ctx.fillStyle = accentColor;
        ctx.fill();
    };

    const [winner, setWinner] = useState<string | null>(null);

    const spin = () => {
        if (isSpinning || items.length === 0) return;

        setIsSpinning(true);
        setWinner(null);
        const spins = 5 + Math.random() * 5;
        const extraRotation = Math.random() * 2 * Math.PI;
        const totalRotation = spins * 2 * Math.PI + extraRotation;
        
        const startRotation = rotation;
        const startTime = performance.now();
        const duration = 3000; // 3 seconds

        const animate = (currentTime: number) => {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            
            // Ease out cubic
            const easeOut = (t: number) => 1 - Math.pow(1 - t, 3);
            const currentRotation = startRotation + totalRotation * easeOut(progress);
            
            setRotation(currentRotation);

            if (progress < 1) {
                requestAnimationFrame(animate);
            } else {
                setIsSpinning(false);
                // Calculate winner
                const finalRotation = currentRotation % (2 * Math.PI);
                const winningAngle = (2 * Math.PI - finalRotation) % (2 * Math.PI);
                const sliceAngle = (2 * Math.PI) / items.length;
                const winnerIndex = Math.floor(winningAngle / sliceAngle);
                setWinner(items[winnerIndex].option);
            }
        };

        requestAnimationFrame(animate);
    };

    return (
        <div className="wheel-container">
            <div className="wheel-wrapper">
                <canvas 
                    ref={canvasRef} 
                    width={500} 
                    height={500} 
                    style={{ width: '100%', height: '100%' }}
                />
            </div>
            
            {winner && (
                <div style={{ fontSize: '1.5rem', fontWeight: 'bold', color: 'var(--accent)' }}>
                    Winner: {winner}!
                </div>
            )}

            <button 
                className="spin-btn" 
                onClick={spin}
                disabled={isSpinning}
            >
                {isSpinning ? 'Spinning...' : 'SPIN'}
            </button>
        </div>
    );
};

export default Wheel;
