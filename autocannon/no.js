const autocannon = require("autocannon");

autocannon(
  {
    url: "http://localhost:1337/movies",
    connections: 100,
    duration: 10,
    pipelining: 10,
    method: "GET",
  },
  onDone
);

function onDone(err, stats) {
  if (err) {
    console.error("Benchmark failed:", err);
    return;
  }

  const toMB = (bytes) => (bytes / (1024 * 1024)).toFixed(2);
  const pad = (label, len = 15) => label.padEnd(len, " ");

  console.log("\n" + "🏁 Benchmark Results\n".toUpperCase());

  console.log(`${pad("URL:")} ${stats.url}`);
  console.log(`${pad("Connections:")} ${stats.connections}`);
  console.log(`${pad("Duration:")} ${stats.duration}s`);
  console.log(`${pad("Pipelining:")} ${stats.pipelining}`);

  console.log("\n📊 Stats:");
  console.log(
    `${pad("Requests/sec:")} ${Math.round(stats.requests.average)} avg, ${
      stats.requests.max
    } max`
  );
  console.log(
    `${pad("Latency (ms):")} ${Math.round(stats.latency.average)} avg, ${
      stats.latency.max
    } max`
  );
  console.log(
    `${pad("Throughput:")} ${toMB(stats.throughput.average)} MB/sec (avg)`
  );

  console.log("\n📈 Totals:");
  console.log(`${pad("Total Requests:")} ${stats.requests.total}`);
  console.log(`${pad("Total Bytes:")} ${toMB(stats.throughput.total)} MB`);
}
