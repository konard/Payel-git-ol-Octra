import { useEffect, useRef, useState } from 'react';

interface TypewriterTextProps {
  text: string;
  // When false the full text is shown immediately (e.g. messages restored from
  // history or the user's own messages). Defaults to true.
  animate?: boolean;
  // Characters revealed per tick. Higher = faster typing.
  speed?: number;
  // Called on every reveal step so the chat can keep scrolling to the bottom
  // while the answer "types" itself out.
  onTick?: () => void;
}

// TypewriterText reveals its text character-by-character so the assistant's
// answers appear to be typed out in real time, like a normal chat, instead of
// popping in all at once (issue #70). Once a message has finished animating it
// stays fully visible; changing the `text` prop restarts the animation.
export function TypewriterText({ text, animate = true, speed = 2, onTick }: TypewriterTextProps) {
  const [count, setCount] = useState(animate ? 0 : text.length);
  const onTickRef = useRef(onTick);
  onTickRef.current = onTick;

  useEffect(() => {
    if (!animate) {
      setCount(text.length);
      return;
    }

    setCount(0);
    const step = Math.max(1, speed);
    const timer = setInterval(() => {
      setCount((prev) => {
        const next = Math.min(text.length, prev + step);
        onTickRef.current?.();
        if (next >= text.length) {
          clearInterval(timer);
        }
        return next;
      });
    }, 18);

    return () => clearInterval(timer);
  }, [text, animate, speed]);

  const visible = animate ? text.slice(0, count) : text;
  const isTyping = animate && count < text.length;

  return (
    <span className="whitespace-pre-wrap">
      {visible}
      {isTyping && (
        <span
          className="ml-0.5 inline-block h-4 w-[2px] translate-y-0.5 animate-pulse bg-current align-middle"
          aria-hidden="true"
        />
      )}
    </span>
  );
}
