# OCTRA C# Client

## 1. Console Client (.NET)

```csharp
using System;
using System.Net.WebSockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

class OctraClient
{
    private string taskId;
    private int backoff = 1000;

    public async Task ConnectAsync()
    {
        string uri = "wss://your-domain.com/ws";

        while (true)
        {
            using var ws = new ClientWebSocket();
            try
            {
                await ws.ConnectAsync(new Uri(uri), CancellationToken.None);

                if (!string.IsNullOrEmpty(taskId))
                    await SendAsync(ws, $"{{\"type\":\"resume\",\"taskId\":\"{taskId}\"}}");
                else
                    await SendAsync(ws, "{\"type\":\"create\",\"title\":\"C# Task\"}");

                var buffer = new byte[4096];
                while (ws.State == WebSocketState.Open)
                {
                    var result = await ws.ReceiveAsync(buffer, CancellationToken.None);
                    var msg = Encoding.UTF8.GetString(buffer, 0, result.Count);
                    Console.WriteLine("Update: " + msg);
                }
            }
            catch
            {
                await Task.Delay(backoff);
                backoff = Math.Min(backoff * 2, 30000);
            }
        }
    }

    private async Task SendAsync(ClientWebSocket ws, string message)
    {
        var bytes = Encoding.UTF8.GetBytes(message);
        await ws.SendAsync(bytes, WebSocketMessageType.Text, true, CancellationToken.None);
    }
}
```

---

## 2. ASP.NET Core Background Service

```csharp
using Microsoft.Extensions.Hosting;
using System.Net.WebSockets;
using System.Text;

public class OctraBackgroundService : BackgroundService
{
    private string taskId;
    private int backoff = 1000;

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                using var ws = new ClientWebSocket();
                await ws.ConnectAsync(new Uri("wss://your-domain.com/ws"), stoppingToken);

                if (!string.IsNullOrEmpty(taskId))
                    await SendAsync(ws, $"{{\"type\":\"resume\",\"taskId\":\"{taskId}\"}}");
                else
                    await SendAsync(ws, "{\"type\":\"create\",\"title\":\"ASP.NET Task\"}");

                var buffer = new byte[4096];
                while (ws.State == WebSocketState.Open && !stoppingToken.IsCancellationRequested)
                {
                    var result = await ws.ReceiveAsync(buffer, stoppingToken);
                    var msg = Encoding.UTF8.GetString(buffer, 0, result.Count);
                    Console.WriteLine("Update: " + msg);
                }
            }
            catch
            {
                await Task.Delay(backoff, stoppingToken);
                backoff = Math.Min(backoff * 2, 30000);
            }
        }
    }

    private async Task SendAsync(ClientWebSocket ws, string message)
    {
        var bytes = Encoding.UTF8.GetBytes(message);
        await ws.SendAsync(bytes, WebSocketMessageType.Text, true, CancellationToken.None);
    }
}
```