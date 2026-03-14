import { AbsoluteFill, Series, Audio, staticFile } from "remotion";
import type { RemotionProps } from "../types";
import { PanelSlide } from "./PanelSlide";

// ShortDrama is the main composition component.
// It uses <Series> to play panels one after another without manual offset calculation.
// Duration is driven dynamically by calculateMetadata in Root.tsx.
export const ShortDrama: React.FC<RemotionProps> = ({ panels, fps, bgm_url }) => {
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
      {bgm_url && <Audio src={staticFile(bgm_url)} loop volume={0.6} />}
      <Series>
        {panels.map((panel, i) => {
          const durationInFrames = Math.max(
            1,
            Math.round(panel.duration_sec * fps)
          );
          return (
            // premountFor=1*fps ensures smooth transitions (Remotion best practice)
            <Series.Sequence
              key={`${panel.scene_number}-${panel.panel_number}-${i}`}
              durationInFrames={durationInFrames}
              premountFor={fps}
            >
              <PanelSlide panel={panel} />
            </Series.Sequence>
          );
        })}
      </Series>
    </AbsoluteFill>
  );
};
