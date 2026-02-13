import { spawn } from "node:child_process";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const sidecarDir = path.resolve(__dirname, "..");

const server = spawn(process.execPath, ["src/server.mjs"], {
  cwd: sidecarDir,
  stdio: ["ignore", "pipe", "pipe"],
});

let readyResolve;
let readyReject;
const ready = new Promise((resolve, reject) => {
  readyResolve = resolve;
  readyReject = reject;
});

const readyTimeout = setTimeout(() => {
  readyReject(new Error("sidecar did not become ready within 5 seconds"));
}, 5000);

server.stdout.on("data", (chunk) => {
  const text = chunk.toString();
  process.stdout.write(`[sidecar] ${text}`);
  if (text.includes("forgemate-sidecar listening on")) {
    clearTimeout(readyTimeout);
    readyResolve();
  }
});

server.stderr.on("data", (chunk) => {
  process.stderr.write(`[sidecar] ${chunk.toString()}`);
});

server.on("exit", (code) => {
  if (code !== 0) {
    clearTimeout(readyTimeout);
    readyReject(new Error(`sidecar exited before ready (code=${code})`));
  }
});

try {
  await ready;
  const smokeCode = await runSmokeClient(sidecarDir);
  await shutdownServer(server);
  process.exit(smokeCode);
} catch (err) {
  console.error("smoke-e2e failed", err);
  await shutdownServer(server);
  process.exit(1);
}

function runSmokeClient(cwd) {
  return new Promise((resolve) => {
    const client = spawn(process.execPath, ["scripts/smoke-run.mjs"], {
      cwd,
      stdio: "inherit",
    });

    client.on("exit", (code) => {
      resolve(code ?? 1);
    });
  });
}

function shutdownServer(proc) {
  return new Promise((resolve) => {
    if (proc.killed || proc.exitCode !== null) {
      resolve();
      return;
    }

    const forceTimer = setTimeout(() => {
      proc.kill("SIGKILL");
    }, 3000);

    proc.once("exit", () => {
      clearTimeout(forceTimer);
      resolve();
    });

    proc.kill("SIGTERM");
  });
}
