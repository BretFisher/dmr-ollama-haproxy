# Ollama API proxy for Docker Model Runner

THIS IS VERY ALPHA with a hack to get DMR working in VSCode Copilot.

**The Problem:** You want to use Docker Desktops new Docker Model Runner (DMR) server, but all your tools are configured to use the Ollama API, which is partially based on OpenAI's API and partially on Ollama's own API. Turns out DMR and Ollama use similar OpenAI endpoints, but their model-listing endpoints are different.

If the tool that supports using Ollama only uses the `/v1/` OpenAI API paths, then simply pointing the tool to `http://localhost:12434/engines/v1/` from the host OS or `http://model-runner.docker.internal/engines/v1/` in a container will work. But, if the tool needs to use the custom `/api/` Ollama paths, DMR doesn't respond to those calls correctly. DMR has its own `/models/` endpoint with a different JSON output.

> **This repo tries to workaround that problem by providing a proxy in front of DMR that looks a bit more like Ollama**

## What this tool does

- Runs from Docker Compose with two services:
  1. A one-shot JSON transformer to turn the DMR models into Ollama API format (`models.json`).
  2. Start HAProxy to listen on Ollama's port `localhost:11434`.
- HAproxy will:
  1. Proxy `/v1/` requests to DMR without any changes.
  2. Responds to `GET /api/tags` with the static JSON file (`models.json`) created on Compose startup.
  3. Responds to `POST /api/show` with a static JSON file (`model.json`) I hand-edited to be generic.
- The fancy part: On Compose start, it'll build and use a `dmr-models-convert` container to get the `/models/` list from the DMR API and transform that into the Ollama JSON equivalent. It then saves that JSON file to the host working directory where HAProxy can get it.
- The Compose file uses the new `depends_on` feature `condition: service_completed_successfully` to hold HAProxy starting until `dmr-models-convert` finishes.

## WARNINGS

- This has only been tested with VS Code's Ollama support.
- Not all model data is exactly the same after conversion from DMR to Ollama format. (model family names might be off, etc.)
- The `/api/show` API call always returns the same generic JSON (`./model.json`). It didn't seem to matter in VS Code (and no way yet to see why VS Code even needs model details as seen in `model.json`) but other tools may not work because of this limitation.
- Other issues probably exist around request size limits (large context windows) or other APIs paths that should be proxied.

## Usage

Clone this repo to your machine and compose up:

```bash
gh repo clone bretfisher/dmr-ollama-haproxy
cd dmr-ollama-haproxy
docker compose up
```

If you add a new DMR model, restart compose to get an updated `models.json` built.

## Models API

Just for comparison, here's a sample of the Ollama response to `/api/tags`, with more in `./example-json`:

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
