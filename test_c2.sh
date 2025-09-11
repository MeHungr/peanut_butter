#!/usr/bin/env bash

# Requires: curl, jq

BASE_URL="http://localhost:8080"

echo "===================="
echo "1. Register a valid agent"
echo "===================="
curl -s -X POST "$BASE_URL/register" \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "agent123"}' 

echo
echo "===================="
echo "2. Register an agent with missing agent_id"
echo "===================="
curl -s -X POST "$BASE_URL/register" \
  -H "Content-Type: application/json" \
  -d '{}' 

echo
echo "===================="
echo "3. Enqueue a normal task"
echo "===================="
curl -s -X POST "$BASE_URL/enqueue" \
  -H "Content-Type: application/json" \
  -d '{
        "agent_id": "agent123",
        "task": {"id": "task1", "type": "echo", "completed": false}
      }' 

echo
echo "===================="
echo "4. Enqueue a task for a non-existent agent"
echo "===================="
curl -s -X POST "$BASE_URL/enqueue" \
  -H "Content-Type: application/json" \
  -d '{
        "agent_id": "fakeagent",
        "task": {"id": "taskX", "type": "echo", "completed": false}
      }' 

echo
echo "===================="
echo "5. Request a task for a non-existent agent"
echo "===================="
curl -s -X POST "$BASE_URL/task" \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "fakeagent"}' -w "HTTP Status: %{http_code}\n" 

echo
echo "===================="
echo "6. Request a task when all tasks are completed"
echo "===================="
# Mark task1 as completed
curl -s -X POST "$BASE_URL/result" \
  -H "Content-Type: application/json" \
  -d '{
        "agent_id": "agent123",
        "task_id": "task1",
        "output": "hello"
      }' 

# Request next task
curl -s -X POST "$BASE_URL/task" \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "agent123"}' -w "HTTP Status: %{http_code}\n" 

echo
echo "===================="
echo "7. Send result for a non-existent agent"
echo "===================="
curl -s -X POST "$BASE_URL/result" \
  -H "Content-Type: application/json" \
  -d '{
        "agent_id": "fakeagent",
        "task_id": "task1",
        "output": "hello"
      }' -w "HTTP Status: %{http_code}\n" 

echo
echo "===================="
echo "8. Send result for a non-existent task"
echo "===================="
curl -s -X POST "$BASE_URL/result" \
  -H "Content-Type: application/json" \
  -d '{
        "agent_id": "agent123",
        "task_id": "taskX",
        "output": "hello"
      }' -w "HTTP Status: %{http_code}\n" 

echo
echo "===================="
echo "9. Enqueue duplicate task IDs (should error if enforced)"
echo "===================="
curl -s -X POST "$BASE_URL/enqueue" \
  -H "Content-Type: application/json" \
  -d '{
        "agent_id": "agent123",
        "task": {"id": "task1", "type": "echo", "completed": false}
      }' 

echo
echo "===================="
echo "10. Enqueue a task with missing fields"
echo "===================="
curl -s -X POST "$BASE_URL/enqueue" \
  -H "Content-Type: application/json" \
  -d '{
        "agent_id": "agent123",
        "task": {}
      }' 

echo
echo "===================="
echo "All tests completed"
echo "===================="
