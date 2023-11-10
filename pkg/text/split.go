package text

import (
	"github.com/pkoukk/tiktoken-go"
)

const defaultChunkTokenSize = 500

// TokenizerSplitterInput defines the input for the tokenizer splitter task
type TokenizerSplitterInput struct {
	// Text: Text to split
	Text string `json:"text"`
	// Model: ID of the model to use for tokenization
	Model string `json:"model"`
	// ChunkTokenSize: Number of tokens per text chunk
	ChunkTokenSize *int `json:"chunk_token_size,omitempty"`
}

// TokenizerSplitterOutput defines the output for the tokenizer splitter task
type TokenizerSplitterOutput struct {
	// Texts: Text chunks
	TokenCount int      `json:"token_count"`
	TextChunks []string `json:"text_chunks"`
	ChunkNum   int      `json:"chunk_num"`
}

// SplitTextIntoChunks splits text into text chunks based on token size
func splitTextIntoChunks(input TokenizerSplitterInput) (TokenizerSplitterOutput, error) {
	output := TokenizerSplitterOutput{}

	if input.ChunkTokenSize == nil || *input.ChunkTokenSize <= 0 {
		input.ChunkTokenSize = new(int)
		*input.ChunkTokenSize = defaultChunkTokenSize
	}

	tkm, err := tiktoken.EncodingForModel(input.Model)
	if err != nil {
		return output, err
	}

	token := tkm.Encode(input.Text, nil, nil)
	output.TokenCount = len(token)
	for start := 0; start < len(token); start += *input.ChunkTokenSize {
		end := min(start+*input.ChunkTokenSize, len(token))
		output.TextChunks = append(output.TextChunks, tkm.Decode(token[start:end]))
	}
	output.ChunkNum = len(output.TextChunks)
	return output, nil
}
