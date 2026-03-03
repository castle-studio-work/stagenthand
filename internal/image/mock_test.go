package image_test

import (
	"context"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/image"
	"github.com/stretchr/testify/assert"
)

func TestMockClient_GenerateImage(t *testing.T) {
	mock := &image.MockClient{}
	
	bytes, err := mock.GenerateImage(context.Background(), "a brave hero", []string{})
	assert.NoError(t, err)
	assert.NotEmpty(t, bytes)
	assert.Equal(t, 1, mock.CallCount)
}
