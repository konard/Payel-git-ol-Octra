/**
 * Lefine.pro ActivityPub Integration Service
 * 
 * This service handles ActivityPub protocol communication with Lefine.pro
 * - Subscribes to /outbox to receive tasks
 * - Sends completed tasks (zip) to /inbox
 * - Integrates with existing workflow system
 */

const LEFINE_PRO_URL = 'https://exchange.lefine.pro';

export interface ActivityPubActor {
  id: string;
  type: string;
  name: string;
  inbox: string;
  outbox: string;
  preferredUsername: string;
}

export interface ActivityPubActivity {
  '@context': string | string[];
  id: string;
  type: string;
  actor: string;
  object: any;
  to?: string | string[];
  cc?: string | string[];
  published?: string;
}

export interface ActivityPubOutboxItem {
  '@context': string;
  id: string;
  type: 'OrderedCollection';
  totalItems: number;
  orderedItems: ActivityPubActivity[];
}

export interface TaskFromLefine {
  id: string;
  title: string;
  description: string;
  workflowId?: string;
  metadata?: Record<string, any>;
}

/**
 * Get ActivityPub WebFinger for a Lefine.pro instance
 */
export async function getLefineActor(baseUrl: string = LEFINE_PRO_URL): Promise<ActivityPubActor | null> {
  try {
    // Try to get the actor from the well-known URL
    const response = await fetch(`${baseUrl}/actor.json`, {
      headers: {
        'Accept': 'application/activity+json',
      },
    });

    if (!response.ok) {
      console.warn('Failed to fetch Lefine actor, using default');
      return createDefaultActor(baseUrl);
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching Lefine actor:', error);
    return createDefaultActor(baseUrl);
  }
}

function createDefaultActor(baseUrl: string): ActivityPubActor {
  return {
    id: `${baseUrl}/actor`,
    type: 'Application',
    name: 'CrewAI',
    inbox: `${baseUrl}/inbox`,
    outbox: `${baseUrl}/outbox`,
    preferredUsername: 'crewai',
  };
}

/**
 * Fetch tasks from Lefine.pro outbox
 * Returns an array of tasks that can be processed
 */
export async function fetchTasksFromOutbox(
  outboxUrl: string = `${LEFINE_PRO_URL}/outbox`
): Promise<TaskFromLefine[]> {
  try {
    const response = await fetch(outboxUrl, {
      headers: {
        'Accept': 'application/activity+json, application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch outbox: ${response.status} ${response.statusText}`);
    }

    const outbox: ActivityPubOutboxItem = await response.json();
    
    // Convert ActivityPub activities to our Task format
    const tasks: TaskFromLefine[] = outbox.orderedItems
      .filter((item) => item.type === 'Create' || item.type === 'Note')
      .map((item) => ({
        id: item.id,
        title: item.object?.name || item.object?.content || 'Untitled Task',
        description: item.object?.content || '',
        workflowId: item.object?.workflowId,
        metadata: item.object,
      }));

    return tasks;
  } catch (error) {
    console.error('Error fetching tasks from outbox:', error);
    return [];
  }
}

/**
 * Send completed task (zip) to Lefine.pro inbox
 * 
 * @param inboxUrl - The inbox endpoint
 * @param taskResult - The completed task result with zip
 * @param apiKey - API key for authentication
 */
export async function sendTaskToInbox(
  taskResult: {
    taskId: string;
    zipUrl: string;
    zipData?: Blob;
    status: string;
    metadata?: Record<string, any>;
  },
  inboxUrl: string = `${LEFINE_PRO_URL}/inbox`,
  apiKey?: string
): Promise<boolean> {
  try {
    // Create ActivityPub Create activity with the task result
    const activity: ActivityPubActivity = {
      '@context': 'https://www.w3.org/ns/activitystreams',
      id: `${inboxUrl}/activities/${Date.now()}`,
      type: 'Create',
      actor: 'crewai',
      object: {
        type: 'Document',
        id: taskResult.taskId,
        name: `Completed Task: ${taskResult.taskId}`,
        content: `Task completed with status: ${taskResult.status}`,
        url: taskResult.zipUrl,
        status: taskResult.status,
        metadata: taskResult.metadata,
        published: new Date().toISOString(),
      },
      published: new Date().toISOString(),
    };

    const headers: Record<string, string> = {
      'Content-Type': 'application/activity+json',
      'Accept': 'application/activity+json, application/json',
    };

    if (apiKey) {
      headers['Authorization'] = `Bearer ${apiKey}`;
    }

    const response = await fetch(inboxUrl, {
      method: 'POST',
      headers,
      body: JSON.stringify(activity),
    });

    if (!response.ok) {
      throw new Error(`Failed to send to inbox: ${response.status} ${response.statusText}`);
    }

    console.log('Task sent to Lefine.pro inbox successfully');
    return true;
  } catch (error) {
    console.error('Error sending task to inbox:', error);
    return false;
  }
}

/**
 * Subscribe to Lefine.pro outbox for real-time updates
 * Uses polling since ActivityPub doesn't have native push
 * 
 * @param callback - Function to call when new tasks are found
 * @param intervalMs - Polling interval in milliseconds
 * @param outboxUrl - The outbox endpoint
 */
export function subscribeToOutbox(
  callback: (tasks: TaskFromLefine[]) => void,
  intervalMs: number = 30000, // 30 seconds default
  outboxUrl: string = `${LEFINE_PRO_URL}/outbox`
): () => void {
  let lastTaskIds: string[] = [];
  let isRunning = true;

  const poll = async () => {
    if (!isRunning) return;

    try {
      const tasks = await fetchTasksFromOutbox(outboxUrl);
      
      // Only call callback if there are new tasks
      const newTasks = tasks.filter(
        (task) => !lastTaskIds.includes(task.id)
      );

      if (newTasks.length > 0) {
        console.log(`Found ${newTasks.length} new tasks from Lefine.pro`);
        callback(newTasks);
        
        // Update last task ids
        lastTaskIds = tasks.map((t) => t.id);
      }
    } catch (error) {
      console.error('Error polling outbox:', error);
    }

    // Schedule next poll
    if (isRunning) {
      setTimeout(poll, intervalMs);
    }
  };

  // Start polling
  poll();

  // Return unsubscribe function
  return () => {
    isRunning = false;
  };
}

/**
 * Process a task from Lefine.pro through the local workflow
 * This integrates with the existing CrewAI workflow system
 */
export async function processLefineTask(
  task: TaskFromLefine,
  workflowId: string,
  apiKey: string
): Promise<string | null> {
  try {
    // Here we would integrate with the existing task creation flow
    // For now, we'll just log the task details
    console.log('Processing Lefine task through workflow:', {
      taskId: task.id,
      title: task.title,
      workflowId,
    });

    // TODO: Integrate with existing task store to create and send the task
    // through the selected workflow. This would involve:
    // 1. Creating a task via the API gateway
    // 2. Connecting to the WebSocket
    // 3. Sending the task description through the workflow
    // 4. Monitoring for completion
    // 5. Sending the result back to Lefine.pro inbox

    return task.id;
  } catch (error) {
    console.error('Error processing Lefine task:', error);
    return null;
  }
}

/**
 * Initialize Lefine.pro integration
 * Sets up the subscription to outbox and processes tasks through the workflow
 */
export function initializeLefineIntegration(
  workflowId: string,
  apiKey: string,
  outboxUrl: string = `${LEFINE_PRO_URL}/outbox`,
  inboxUrl: string = `${LEFINE_PRO_URL}/inbox`
): { unsubscribe: () => void } {
  console.log('Initializing Lefine.pro integration...');

  // Subscribe to outbox for new tasks
  const unsubscribe = subscribeToOutbox(
    async (tasks) => {
      for (const task of tasks) {
        // Process each task through the workflow
        const processedId = await processLefineTask(task, workflowId, apiKey);
        
        if (processedId) {
          console.log(`Task ${processedId} queued for processing`);
        }
      }
    },
    30000, // Poll every 30 seconds
    outboxUrl
  );

  return { unsubscribe };
}
