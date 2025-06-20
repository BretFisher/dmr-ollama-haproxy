# JSON converter for DMR model output to Ollama

Problem: You want to use Docker Desktops new Docker Model Runner (DMR)` server, but all your tools are configured to use the Ollama API, which is partially based on OpenAI's API and partially on Ollama's own API. Turns out they both use similar OpenAI endpoints, but their model-listing endpoints are different.

This Golang CLI app provides a simple converter that pulls from
DMR's API at `http://localhost:12434/models` and transforms that JSON response into the Ollama API format from `/api/tags`. It saves this file to an output location, which can then be served by a webserver to emulate the Ollama API.

## Usage

To convert models from a running DMR server:

```bash
# Print to stdout (uses default DMR server)
./dmr-models-convert
```

### Command Options

- `--output` / `-o`: Output file path for converted JSON (optional, prints to stdout if not specified)
- `--dmr` / `-d`: DMR server URL (optional, defaults to http://localhost:12434/models)


### Building the Application

```bash
# Build the application with go
go build -o dmr-models-convert
```

## Models API

Here's a sample of the Ollama response to `/api/tags`:

```json
{
  "models": [
    {
      "name": "smollm2:360m",
      "model": "smollm2:360m",
      "modified_at": "2025-06-19T22:47:03.926871948-04:00",
      "size": 725566512,
      "digest": "297281b699fc51376006233ca400cd664c4f7b80ed88a47ef084f1e4b089803b",
      "details": {
        "parent_model": "",
        "format": "gguf",
        "family": "llama",
        "families": [
          "llama"
        ],
        "parameter_size": "361.82M",
        "quantization_level": "F16"
      }
    }
  ]
}
```

Here's a sample response from DMR `/models`:

```json
[
  {
    "id": "sha256:020ef929a2866cc4079bf477583c23dc1432e37b9e73b3c20de51a3720b90ac7",
    "tags": [
      "ai/smollm2:360M-F16"
    ],
    "created": 1745698622,
    "config": {
      "format": "gguf",
      "quantization": "F16",
      "parameters": "361.82 M",
      "architecture": "llama",
      "size": "690.24 MiB"
    }
  }
]
```
