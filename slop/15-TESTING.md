# Testing

I see the following errors:

```
2026/06/16 19:08:44 Synchronization engine is running. Press Ctrl+C to stop.
2026/06/16 19:08:44 [SyncEngine] Triggering initial Signal receive catch-up...
2026/06/16 19:08:44 [signal-cli stderr] ERROR SignalJsonRpcDispatcherHandler - Command execution failed
2026/06/16 19:08:44 [signal-cli stderr] java.lang.NullPointerException
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.commands.ReceiveCommand.handleCommand(ReceiveCommand.java:112)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.commands.ReceiveCommand.handleCommand(ReceiveCommand.java:33)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.jsonrpc.SignalJsonRpcCommandHandler$CommandRunnerImpl.handleCommand(SignalJsonRpcCommandHandler.java:182)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.jsonrpc.SignalJsonRpcCommandHandler.parseParamsAndRunCommand(SignalJsonRpcCommandHandler.java:309)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.jsonrpc.SignalJsonRpcCommandHandler.runCommand(SignalJsonRpcCommandHandler.java:243)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.jsonrpc.SignalJsonRpcCommandHandler.handleRequest(SignalJsonRpcCommandHandler.java:103)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.jsonrpc.SignalJsonRpcDispatcherHandler.lambda$handleConnection$5(SignalJsonRpcDispatcherHandler.java:209)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.jsonrpc.JsonRpcReader.handleRequest(JsonRpcReader.java:134)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.jsonrpc.JsonRpcReader.handleMessage(JsonRpcReader.java:85)
2026/06/16 19:08:44 [signal-cli stderr] 	at org.asamk.signal.jsonrpc.JsonRpcReader.lambda$readMessages$0(JsonRpcReader.java:72)
2026/06/16 19:08:44 [signal-cli stderr] 	at java.base@25.0.2/java.util.concurrent.Executors$RunnableAdapter.call(Executors.java:545)
2026/06/16 19:08:44 [signal-cli stderr] 	at java.base@25.0.2/java.util.concurrent.FutureTask.run(FutureTask.java:328)
2026/06/16 19:08:44 [signal-cli stderr] 	at java.base@25.0.2/java.util.concurrent.ThreadPerTaskExecutor$ThreadBoundFuture.run(ThreadPerTaskExecutor.java:323)
2026/06/16 19:08:44 [signal-cli stderr] 	at java.base@25.0.2/java.lang.Thread.runWith(Thread.java:1487)
2026/06/16 19:08:44 [signal-cli stderr] 	at java.base@25.0.2/java.lang.VirtualThread.run(VirtualThread.java:456)
2026/06/16 19:08:44 [signal-cli stderr] 	at java.base@25.0.2/java.lang.VirtualThread$VThreadContinuation$1.run(VirtualThread.java:248)
2026/06/16 19:08:44 [SyncEngine] Initial Signal receive trigger failed: signal-cli error -32603: null (NullPointerException)
```

Fix it.
