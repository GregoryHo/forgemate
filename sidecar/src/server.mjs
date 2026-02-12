import fs from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import grpc from "@grpc/grpc-js";
import protoLoader from "@grpc/proto-loader";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const PROTO_PATH = path.resolve(__dirname, "../../proto/runtime/v1/agent_runtime.proto");

const runtime = loadRuntime();
const activeRuns = new Map();
const bindAddress = resolveBindAddress();

const server = new grpc.Server();
server.addService(runtime.AgentRuntime.service, {
  Health: handleHealth,
  Run: handleRun,
  Abort: handleAbort,
  Wait: handleWait,
  AuthProbe: handleAuthProbe,
});

cleanupUnixSocket(bindAddress);

server.bindAsync(bindAddress, grpc.ServerCredentials.createInsecure(), (err) => {
  if (err) {
    console.error("failed to bind sidecar server", err);
    process.exit(1);
  }
  console.log(`forgemate-sidecar listening on ${bindAddress}`);
});

process.on("SIGINT", () => shutdown("SIGINT"));
process.on("SIGTERM", () => shutdown("SIGTERM"));

function loadRuntime() {
  const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true,
  });

  const loaded = grpc.loadPackageDefinition(packageDefinition);
  return loaded.forgemate.runtime.v1;
}

function resolveBindAddress() {
  if (process.env.FORGEMATE_SIDECAR_ADDR) {
    return process.env.FORGEMATE_SIDECAR_ADDR;
  }
  const socketPath = process.env.FORGEMATE_SIDECAR_SOCKET || "/tmp/forgemate-agent-runtime.sock";
  return `unix:${socketPath}`;
}

function cleanupUnixSocket(address) {
  const socketPath = parseUnixSocketPath(address);
  if (!socketPath) {
    return;
  }

  try {
    fs.rmSync(socketPath, { force: true });
  } catch (err) {
    console.warn(`unable to remove stale socket ${socketPath}`, err);
  }
}

function parseUnixSocketPath(address) {
  if (!address.startsWith("unix:")) {
    return "";
  }

  const value = address.slice("unix:".length);
  if (!value) {
    return "";
  }

  if (value.startsWith("//")) {
    return value.slice(1);
  }

  return value;
}

function shutdown(signalName) {
  console.log(`forgemate-sidecar shutdown requested (${signalName})`);
  server.tryShutdown(() => {
    cleanupUnixSocket(bindAddress);
    process.exit(0);
  });

  setTimeout(() => {
    server.forceShutdown();
    cleanupUnixSocket(bindAddress);
    process.exit(1);
  }, 3000).unref();
}

function handleHealth(_, callback) {
  callback(null, {
    ok: true,
    state: "running",
    runtime: "node-sidecar",
  });
}

function handleRun(call) {
  let outboundSeq = 1;
  let terminal = false;

  const writeFrame = (frame) => {
    call.write({ seq: outboundSeq, ...frame });
    outboundSeq += 1;
  };

  call.on("data", (frame) => {
    if (frame.ping) {
      writeFrame({ pong: { unixMs: Date.now() } });
      return;
    }

    if (frame.abortSignal) {
      const runId = frame.abortSignal.runId || "unknown";
      activeRuns.delete(runId);
      writeFrame({ runFailed: { runId, code: "aborted", message: frame.abortSignal.reason || "aborted" } });
      terminal = true;
      call.end();
      return;
    }

    if (frame.toolResult) {
      writeFrame({
        toolProgress: {
          runId: frame.toolResult.runId,
          callId: frame.toolResult.toolCallId,
          status: "result-received",
        },
      });
      return;
    }

    if (!frame.runStart) {
      return;
    }

    const runId = frame.runStart.runId || `run-${Date.now()}`;
    const prompt = frame.runStart.prompt || "";
    const output = prompt ? `sidecar-echo: ${prompt}` : "sidecar-ready";

    activeRuns.set(runId, {
      startedAt: Date.now(),
      sessionKey: frame.runStart.sessionKey,
    });

    writeFrame({ runAccepted: { runId, sidecarRunId: runId } });
    writeFrame({ assistantDelta: { runId, text: output } });
    writeFrame({ runCompleted: { runId, outputText: output } });

    activeRuns.delete(runId);
    terminal = true;
    call.end();
  });

  call.on("end", () => {
    if (!terminal) {
      call.end();
    }
  });

  call.on("error", () => {
    // Stream errors are expected during abrupt client disconnects.
  });
}

function handleAbort(call, callback) {
  const runId = call.request.runId || "";
  const existed = activeRuns.delete(runId);
  callback(null, {
    accepted: true,
    state: existed ? "aborted" : "not-found",
  });
}

function handleWait(call, callback) {
  const runId = call.request.runId || "";
  const terminalState = activeRuns.has(runId) ? "running" : "completed";
  callback(null, {
    runId,
    terminalState,
  });
}

function handleAuthProbe(call, callback) {
  const provider = call.request.provider || "unknown";
  callback(null, {
    provider,
    reachable: true,
    notes: "stub-auth-probe",
  });
}
