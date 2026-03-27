import { useState, useRef, useEffect, useCallback } from 'preact/hooks';
import type { Item } from '../App';

interface WheelProps {
    items: Item[];
    isDarkMode: boolean;
}

const LOGICAL_SIZE = 500;

/** Smoother long deceleration than cubic (settles gently at the end). */
function easeOutQuint(t: number): number {
    return 1 - Math.pow(1 - t, 5);
}

/** Truncate to fit max width (px) using canvas measureText. */
function truncateToWidth(
    ctx: CanvasRenderingContext2D,
    text: string,
    maxWidth: number
): string {
    if (ctx.measureText(text).width <= maxWidth) return text;
    const ellipsis = '…';
    for (let i = text.length - 1; i >= 0; i--) {
        const candidate = text.slice(0, i) + ellipsis;
        if (ctx.measureText(candidate).width <= maxWidth) return candidate;
    }
    return ellipsis;
}

const Wheel = ({ items, isDarkMode }: WheelProps) => {
    const wrapperRef = useRef<HTMLDivElement>(null);
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const rotationRef = useRef(0);
    const rafRef = useRef<number>(0);

    const [isSpinning, setIsSpinning] = useState(false);
    const [winner, setWinner] = useState<string | null>(null);

    const accentColor = '#0071e3';
    const textColor = isDarkMode ? '#f5f5f7' : '#1d1d1f';

    const drawWheel = useCallback(
        (rotation: number) => {
            const canvas = canvasRef.current;
            const wrapper = wrapperRef.current;
            if (!canvas || !wrapper) return;
            const ctx = canvas.getContext('2d');
            if (!ctx) return;

            const displayCssPx = Math.max(
                1,
                Math.floor(wrapper.getBoundingClientRect().width),
            );
            const dpr = window.devicePixelRatio || 1;
            const bitmapSize = Math.max(1, Math.round(displayCssPx * dpr));
            if (canvas.width !== bitmapSize || canvas.height !== bitmapSize) {
                canvas.width = bitmapSize;
                canvas.height = bitmapSize;
            }
            // Draw in LOGICAL_SIZE space; scale to the actual backing-store pixels (sharp on retina).
            const scale = bitmapSize / LOGICAL_SIZE;
            ctx.setTransform(scale, 0, 0, scale, 0, 0);

            const size = LOGICAL_SIZE;
            const center = size / 2;
            const radius = size / 2 - 14;
            const hubR = 22;

            ctx.clearRect(0, 0, size, size);

            if (items.length === 0) {
                ctx.beginPath();
                ctx.arc(center, center, radius * 0.35, 0, 2 * Math.PI);
                ctx.fillStyle = isDarkMode ? '#2c2c2e' : '#e8e8ed';
                ctx.fill();
                ctx.fillStyle = isDarkMode ? '#86868b' : '#6e6e73';
                ctx.font = '600 18px -apple-system, BlinkMacSystemFont, sans-serif';
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';
                ctx.fillText('Add options to spin', center, center);
                return;
            }

            const sliceAngle = (2 * Math.PI) / items.length;

            items.forEach((_, i) => {
                const angle = rotation + i * sliceAngle;
                const hue = (i * 360) / items.length;

                ctx.beginPath();
                ctx.moveTo(center, center);
                ctx.arc(center, center, radius, angle, angle + sliceAngle);
                ctx.closePath();

                const inner = isDarkMode
                    ? `hsl(${hue}, 48%, 38%)`
                    : `hsl(${hue}, 78%, 94%)`;
                const outer = isDarkMode
                    ? `hsl(${hue}, 52%, 24%)`
                    : `hsl(${hue}, 62%, 72%)`;
                const grd = ctx.createRadialGradient(
                    center,
                    center,
                    radius * 0.08,
                    center,
                    center,
                    radius
                );
                grd.addColorStop(0, inner);
                grd.addColorStop(0.55, outer);
                grd.addColorStop(1, isDarkMode ? `hsl(${hue}, 55%, 18%)` : `hsl(${hue}, 55%, 62%)`);

                ctx.fillStyle = grd;
                ctx.fill();

                ctx.strokeStyle = isDarkMode
                    ? 'rgba(255,255,255,0.08)'
                    : 'rgba(255,255,255,0.55)';
                ctx.lineWidth = 1.25;
                ctx.stroke();
            });

            // Outer metal-style rim
            ctx.beginPath();
            ctx.arc(center, center, radius, 0, 2 * Math.PI);
            const rimGrad = ctx.createLinearGradient(0, center - radius, 0, center + radius);
            if (isDarkMode) {
                rimGrad.addColorStop(0, '#5a5a62');
                rimGrad.addColorStop(0.5, '#2e2e33');
                rimGrad.addColorStop(1, '#1a1a1f');
            } else {
                rimGrad.addColorStop(0, '#ffffff');
                rimGrad.addColorStop(0.5, '#d8d8de');
                rimGrad.addColorStop(1, '#a8a8b0');
            }
            ctx.strokeStyle = rimGrad;
            ctx.lineWidth = 10;
            ctx.stroke();

            // Inner subtle ring
            ctx.beginPath();
            ctx.arc(center, center, radius - 6, 0, 2 * Math.PI);
            ctx.strokeStyle = isDarkMode ? 'rgba(255,255,255,0.04)' : 'rgba(255,255,255,0.35)';
            ctx.lineWidth = 1;
            ctx.stroke();

            // Labels: centered along each slice bisector, mid-annulus, sized to slice width
            const textRadius = hubR + (radius - hubR) * 0.62;
            const fontSize = Math.max(
                13,
                Math.min(20, Math.floor(300 / Math.max(items.length * 3.4, 3.4)))
            );
            const maxLabelWidth = Math.max(
                36,
                2 * textRadius * Math.sin(sliceAngle / 2) * 0.92
            );

            items.forEach((item, i) => {
                const angle = rotation + i * sliceAngle;
                ctx.save();
                ctx.translate(center, center);
                ctx.rotate(angle + sliceAngle / 2);
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';
                ctx.font = `600 ${fontSize}px ui-sans-serif, -apple-system, BlinkMacSystemFont, sans-serif`;
                const label = truncateToWidth(ctx, item.option.trim(), maxLabelWidth);
                const strokeColor = isDarkMode
                    ? 'rgba(0,0,0,0.65)'
                    : 'rgba(255,255,255,0.92)';
                ctx.lineWidth = Math.max(2.5, fontSize * 0.2);
                ctx.lineJoin = 'round';
                ctx.miterLimit = 2;
                ctx.strokeStyle = strokeColor;
                ctx.fillStyle = textColor;
                ctx.shadowBlur = 0;
                ctx.strokeText(label, textRadius, 0);
                ctx.fillText(label, textRadius, 0);
                ctx.restore();
            });

            // Center hub
            ctx.beginPath();
            ctx.arc(center, center, hubR, 0, 2 * Math.PI);
            const hubGrad = ctx.createRadialGradient(
                center - 6,
                center - 6,
                2,
                center,
                center,
                hubR
            );
            hubGrad.addColorStop(0, '#4da3ff');
            hubGrad.addColorStop(0.65, accentColor);
            hubGrad.addColorStop(1, '#004999');
            ctx.fillStyle = hubGrad;
            ctx.fill();
            ctx.strokeStyle = 'rgba(255,255,255,0.45)';
            ctx.lineWidth = 2;
            ctx.stroke();

            ctx.beginPath();
            ctx.arc(center - 5, center - 5, hubR * 0.35, 0, 2 * Math.PI);
            ctx.fillStyle = 'rgba(255,255,255,0.22)';
            ctx.fill();

            // Pointer at 3 o'clock — tip sits just outside the rim, aligned with wheel center
            const tipX = center + radius + 5;
            const arrowDepth = 26;
            const arrowHalfH = 15;
            const baseX = tipX - arrowDepth;

            ctx.save();
            ctx.lineJoin = 'round';
            ctx.lineCap = 'round';
            ctx.shadowColor = isDarkMode ? 'rgba(0,0,0,0.45)' : 'rgba(0,0,0,0.22)';
            ctx.shadowBlur = 14;
            ctx.shadowOffsetY = 4;
            ctx.beginPath();
            ctx.moveTo(tipX, center);
            ctx.lineTo(baseX, center - arrowHalfH);
            ctx.lineTo(baseX, center + arrowHalfH);
            ctx.closePath();
            const ptrGrad = ctx.createLinearGradient(baseX - 2, center, tipX, center);
            ptrGrad.addColorStop(0, '#7ec0ff');
            ptrGrad.addColorStop(0.45, '#4a9fff');
            ptrGrad.addColorStop(1, accentColor);
            ctx.fillStyle = ptrGrad;
            ctx.fill();
            ctx.restore();

            ctx.save();
            ctx.lineJoin = 'round';
            ctx.lineCap = 'round';
            ctx.strokeStyle = 'rgba(255,255,255,0.55)';
            ctx.lineWidth = 1.75;
            ctx.beginPath();
            ctx.moveTo(tipX, center);
            ctx.lineTo(baseX, center - arrowHalfH);
            ctx.lineTo(baseX, center + arrowHalfH);
            ctx.closePath();
            ctx.stroke();
            ctx.restore();

            ctx.save();
            ctx.fillStyle = 'rgba(255,255,255,0.35)';
            ctx.beginPath();
            ctx.arc(tipX - 4, center - 4, 3.5, 0, 2 * Math.PI);
            ctx.fill();
            ctx.restore();
        },
        [items, isDarkMode, textColor]
    );

    useEffect(() => {
        drawWheel(rotationRef.current);
    }, [drawWheel, items, isDarkMode]);

    useEffect(() => {
        const wrap = wrapperRef.current;
        if (!wrap || typeof ResizeObserver === 'undefined') return;
        const ro = new ResizeObserver(() => {
            drawWheel(rotationRef.current);
        });
        ro.observe(wrap);
        return () => ro.disconnect();
    }, [drawWheel]);

    useEffect(() => {
        return () => {
            if (rafRef.current) cancelAnimationFrame(rafRef.current);
        };
    }, []);

    const spin = () => {
        if (isSpinning || items.length === 0) return;

        setIsSpinning(true);
        setWinner(null);

        const spins = 6 + Math.random() * 4;
        const extraRotation = Math.random() * 2 * Math.PI;
        const totalRotation = spins * 2 * Math.PI + extraRotation;

        const startRotation = rotationRef.current;
        const startTime = performance.now();
        const duration = 4200;

        const tick = (currentTime: number) => {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            const eased = easeOutQuint(progress);
            rotationRef.current = startRotation + totalRotation * eased;
            drawWheel(rotationRef.current);

            if (progress < 1) {
                rafRef.current = requestAnimationFrame(tick);
            } else {
                setIsSpinning(false);
                const finalRotation = rotationRef.current % (2 * Math.PI);
                const winningAngle = (2 * Math.PI - finalRotation) % (2 * Math.PI);
                const sliceAngle = (2 * Math.PI) / items.length;
                const winnerIndex = Math.floor(winningAngle / sliceAngle);
                setWinner(items[winnerIndex].option);
            }
        };

        rafRef.current = requestAnimationFrame(tick);
    };

    return (
        <div className={`wheel-container${isSpinning ? ' wheel-container--spinning' : ''}`}>
            <div className="wheel-wrapper" ref={wrapperRef}>
                <canvas ref={canvasRef} className="wheel-canvas" />
            </div>

            {winner && (
                <div className="wheel-winner" role="status">
                    <span className="wheel-winner__label">Winner</span>
                    <span className="wheel-winner__name">{winner}</span>
                </div>
            )}

            <button
                type="button"
                className="spin-btn"
                onClick={spin}
                disabled={isSpinning || items.length === 0}
            >
                {isSpinning ? <span className="spin-btn__spinning">Spinning</span> : 'Spin'}
            </button>
        </div>
    );
};

export default Wheel;
