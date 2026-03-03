package image

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

type WorkerResult struct {
	Index int
	Panel domain.Panel
	Err   error
}

type generatorRequest struct {
	Index int
	Panel domain.Panel
}

// GenerateBatch runs image generation over multiple panels concurrently.
func GenerateBatch(ctx context.Context, client Client, panels []domain.Panel, outputDir string, workerCount int) ([]domain.Panel, []error) {
	if workerCount <= 0 {
		workerCount = 3
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return panels, []error{fmt.Errorf("creating output dir: %w", err)}
	}

	var wg sync.WaitGroup
	reqChan := make(chan generatorRequest, len(panels))
	resChan := make(chan WorkerResult, len(panels))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range reqChan {
				imgBytes, err := client.GenerateImage(ctx, req.Panel.Description, req.Panel.CharacterRefs)
				if err != nil {
					req.Panel.ImageURL = "error.png"
					resChan <- WorkerResult{req.Index, req.Panel, fmt.Errorf("panel scene_%d_panel_%d: %w", req.Panel.SceneNumber, req.Panel.PanelNumber, err)}
					continue
				}

				if len(imgBytes) > 0 {
					fileName := fmt.Sprintf("scene_%d_panel_%d.png", req.Panel.SceneNumber, req.Panel.PanelNumber)
					filePath := filepath.Join(outputDir, fileName)
					if err := os.WriteFile(filePath, imgBytes, 0644); err != nil {
						req.Panel.ImageURL = "error.png"
						resChan <- WorkerResult{req.Index, req.Panel, fmt.Errorf("write error panel %d: %w", req.Panel.PanelNumber, err)}
					} else {
						req.Panel.ImageURL = filePath
						resChan <- WorkerResult{req.Index, req.Panel, nil}
					}
				} else {
					req.Panel.ImageURL = "error.png"
					resChan <- WorkerResult{req.Index, req.Panel, nil}
				}
			}
		}()
	}

	for i, p := range panels {
		reqChan <- generatorRequest{Index: i, Panel: p}
	}
	close(reqChan)
	wg.Wait()
	close(resChan)

	outPanels := make([]domain.Panel, len(panels))
	copy(outPanels, panels)
	var errs []error
	for res := range resChan {
		outPanels[res.Index] = res.Panel
		if res.Err != nil {
			errs = append(errs, res.Err)
		}
	}

	return outPanels, errs
}
