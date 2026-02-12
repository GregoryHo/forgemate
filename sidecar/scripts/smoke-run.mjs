import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import grpc from "@grpc/grpc-js";
import protoLoader from "@grpc/proto-loader";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const PROTO_PATH = path.resolve(__dirname, "../../proto/runtime/v1/agent_runtime.proto");
const address = process.env.FORGEMATE_SIDECAR_ADDR || `unix:${process.env.FORGEMATE_SIDECAR_SOCKET || "/tmp/forgemate-agent-runtime.sock"}`;

const runtime = loadRuntime();
const client = new runtime.AgentRuntime(address, grpc.credentials.createInsecure());

try {
  const health = await callUnary((done) => client.Health({}, done));
  console.log("health", JSON.stringify(health));

  const frames = await runSmoke(client);
  console.log("run-frames", JSON.stringify(frames));

  const wait = await callUnary((done) => client.Wait({ runId: "smoke-run-1", timeoutMs: 0 }, done));
  console.log("wait", JSON.stringify(wait));

  const probe = await callUnary((done) => client.AuthProbe({ provider: "openai" }, done));
  console.log("auth-probe", JSON.stringify(probe));
} finally {
  client.close();
}

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

function callUnary(invoke) {
  return new Promise((resolve, reject) => {
    invoke((err, response) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(response);
    });
  });
}

function runSmoke(clientInstance) {
  return new Promise((resolve, reject) => {
    const stream = clientInstance.Run();
    const frames = [];

    stream.on("data", (frame) => {
      frames.push(frame);
      if (frame.runCompleted || frame.runFailed) {
        stream.end();
      }
    });

    stream.on("error", reject);
    stream.on("end", () => resolve(frames));

    stream.write({
      seq: 1,
      runStart: {
        runId: "smoke-run-1",
        sessionKey: "smoke-session",
        prompt: "hello from smoke",
      },
    });
  });
}
