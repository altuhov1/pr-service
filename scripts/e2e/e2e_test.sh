#!/bin/bash
echo "=== E2E TESTING PR REVIEWER SERVICE ==="

BASE_URL="http://localhost:8080"

echo -e "\n1. CREATING TEAMS..."
curl -X POST $BASE_URL/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "team_name": "backend", "is_active": true},
      {"user_id": "u2", "username": "Bob", "team_name": "backend", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "team_name": "backend", "is_active": true}
    ]
  }' && echo -e "\n---"

curl -X POST $BASE_URL/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "frontend", 
    "members": [
      {"user_id": "u4", "username": "David", "team_name": "frontend", "is_active": true},
      {"user_id": "u5", "username": "Eve", "team_name": "frontend", "is_active": true}
    ]
  }' && echo -e "\n---"

echo -e "\n2. CHECKING TEAMS..."
curl -X GET "$BASE_URL/team/get?team_name=backend" && echo -e "\n---"
curl -X GET "$BASE_URL/team/get?team_name=frontend" && echo -e "\n---"

echo -e "\n3. CREATING PR..."
curl -X POST $BASE_URL/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search feature", 
    "author_id": "u1"
  }' && echo -e "\n---"

echo -e "\n4. CHECKING ASSIGNED REVIEWERS..."
curl -X GET "$BASE_URL/users/getReview?user_id=u2" && echo -e "\n---"
curl -X GET "$BASE_URL/users/getReview?user_id=u3" && echo -e "\n---"

echo -e "\n5. DEACTIVATING USER..."
curl -X POST $BASE_URL/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{"user_id": "u2", "is_active": false}' && echo -e "\n---"

echo -e "\n6. REASSIGNING REVIEWER..."
curl -X POST $BASE_URL/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u2"
  }' && echo -e "\n---"

echo -e "\n7. MERGING PR..."
curl -X POST $BASE_URL/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr-1001"}' && echo -e "\n---"

echo -e "\n8. TRYING TO MODIFY MERGED PR (SHOULD FAIL)..."
curl -X POST $BASE_URL/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u3"
  }' && echo -e "\n---"

echo -e "\n9. FINAL CHECK..."
curl -X GET "$BASE_URL/users/getReview?user_id=u3" && echo -e "\n---"

echo "=== E2E TESTING COMPLETED ==="