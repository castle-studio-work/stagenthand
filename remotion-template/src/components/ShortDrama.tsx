import { AbsoluteFill, Series, Audio, staticFile, interpolate, useCurrentFrame, useVideoConfig } from "remotion";
import type { RemotionProps, Directives } from "../types";
import { PanelSlide } from "./PanelSlide";

// Default directives when none provided
const DD: Required<Directives> = {
  bgm_fade_in_sec: 2.0,
  bgm_fade_out_sec: 3.0,
  bgm_volume: 0.6,
  ducking_depth: 0.15,
  ducking_fade_sec: 0.5,
  color_filter: "none",
};

function dd(d?: Directives): Required<Directives> {
  return { ...DD, ...(d ?? {}) };
}

// BGMAudio handles fade-in, fade-out, and auto-ducking of background music.
const BGMAudio: React.FC<{
  bgmUrl: string;
  directives: Required<Directives>;
  panels: RemotionProps["panels"];
  fps: number;
  totalFrames: number;
}> = ({ bgmUrl, directives, panels, fps, totalFrames }) => {
  const frame = useCurrentFrame();

  // 1. Fade envelope
  const fadeInFrames = Math.round(directives.bgm_fade_in_sec * fps);
  const fadeOutFrames = Math.round(directives.bgm_fade_out_sec * fps);

  const fadeIn = interpolate(frame, [0, fadeInFrames], [0, 1], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });
  const fadeOut = interpolate(
    frame,
    [totalFrames - fadeOutFrames, totalFrames],
    [1, 0],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );
  const fadeEnvelope = Math.min(fadeIn, fadeOut);

  // 2. Ducking: lower BGM volume when a panel has audio_url
  let duckFactor = 1.0;
  let accumulatedFrames = 0;
  const duckFadeFrames = Math.round(directives.ducking_fade_sec * fps);

  for (const panel of panels) {
    const panelFrames = Math.round(panel.duration_sec * fps);
    const panelStart = accumulatedFrames;
    const panelEnd = accumulatedFrames + panelFrames;

    if (panel.audio_url) {
      // Inside a voiceover panel: duck the BGM
      const duckIn = interpolate(
        frame,
        [panelStart, panelStart + duckFadeFrames],
        [1.0, directives.ducking_depth / directives.bgm_volume],
        { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
      );
      const duckOut = interpolate(
        frame,
        [panelEnd - duckFadeFrames, panelEnd],
        [directives.ducking_depth / directives.bgm_volume, 1.0],
        { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
      );

      if (frame >= panelStart && frame < panelEnd) {
        duckFactor = Math.min(duckIn, duckOut);
      }
    }
    accumulatedFrames = panelEnd;
  }

  const finalVolume = directives.bgm_volume * fadeEnvelope * duckFactor;

  return <Audio src={staticFile(bgmUrl)} loop volume={finalVolume} />;
};


// ShortDrama is the main composition component.
// It uses <Series> to play panels one after another.
// Duration is driven dynamically by calculateMetadata in Root.tsx.
export const ShortDrama: React.FC<RemotionProps> = ({
  panels,
  fps,
  bgm_url,
  directives: rawDirectives,
}) => {
  const { durationInFrames } = useVideoConfig();
  const dir = dd(rawDirectives);

  if (!panels || panels.length === 0) {
    return (
      <AbsoluteFill
        style={{
          backgroundColor: "#000",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <div style={{ color: "#666", fontFamily: "sans-serif", fontSize: 28 }}>
          No panels provided
        </div>
      </AbsoluteFill>
    );
  }

  return (
    <AbsoluteFill style={{ backgroundColor: "#000" }}>
      {bgm_url && (
        <BGMAudio
          bgmUrl={bgm_url}
          directives={dir}
          panels={panels}
          fps={fps}
          totalFrames={durationInFrames}
        />
      )}
      <Series>
        {panels.map((panel, i) => {
          const durationInFrames = Math.max(
            1,
            Math.round(panel.duration_sec * fps)
          );
          return (
            <Series.Sequence
              key={`${panel.scene_number}-${panel.panel_number}-${i}`}
              durationInFrames={durationInFrames}
              premountFor={fps}
            >
              <PanelSlide panel={panel} colorFilter={dir.color_filter} />
            </Series.Sequence>
          );
        })}
      </Series>
    </AbsoluteFill>
  );
};
