import type React from "react";

interface WatermarkProps {
  /** The text to display in the watermark. */
  text: string;
}

/**
 * A component to display a tiled, rotated watermark over the content.
 * It's designed to be non-intrusive and theme-aware.
 */
const Watermark: React.FC<WatermarkProps> = ({ text }) => {
  // Define the dimensions of the SVG pattern tile.
  // A larger size creates more space between repetitions.
  const svgWidth = 400;
  const svgHeight = 400;

  // Create the SVG string for the watermark pattern.
  // It uses CSS variables for theming, making it adapt to light/dark mode.
  // The text is rotated for a diagonal effect.
  const svgString = `
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="${svgWidth}"
      height="${svgHeight}"
      viewBox="0 0 ${svgWidth} ${svgHeight}"
    >
      <text
        x="50%"
        y="50%"
        text-anchor="middle"
        dominant-baseline="middle"
        transform="rotate(-30, ${svgWidth / 2}, ${svgHeight / 2})"
        style="
          font-size: 16px;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji';
          fill: hsl(var(--foreground));
          opacity: 0.1;
          font-weight: 600;
        "
      >
        ${text}
      </text>
    </svg>
  `;

  // To handle Unicode characters, we URI-encode the SVG string instead of using btoa.
  // This makes it safe for use in a data URL.
  const dataUrl = `url("data:image/svg+xml,${encodeURIComponent(
    svgString.replace(/[\n\s]+/g, " "),
  )}")`;

  return (
    <div
      className="pointer-events-none fixed inset-0 z-30"
      style={{
        backgroundImage: dataUrl,
        backgroundRepeat: "repeat",
      }}
      // Hide from screen readers as it's a purely visual element.
      aria-hidden="true"
    />
  );
};

export default Watermark;
