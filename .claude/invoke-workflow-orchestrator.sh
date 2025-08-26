#!/bin/bash
# Agent Invocation: workflow-orchestrator
# Usage: ./invoke-workflow-orchestrator.sh "Plan implementation for ticket view using TDD approach"

set -e

TASK="${1:-Please provide a task description}"
AGENT_FILE=".claude/agents/workflow-orchestrator.md"

if [ ! -f "$AGENT_FILE" ]; then
    echo "Error: workflow-orchestrator agent not found at $AGENT_FILE"
    exit 1
fi

echo "🎯 Consulting workflow-orchestrator for task: $TASK"
echo "📋 Agent Requirements: $(grep -A5 "When invoked:" $AGENT_FILE | tail -4)"
echo ""
echo "📝 Workflow orchestrator context query:"
echo '{
  "requesting_agent": "workflow-orchestrator", 
  "request_type": "get_workflow_context",
  "payload": {
    "query": "Workflow context needed: process requirements, integration points, error handling needs, performance targets, and compliance requirements.",
    "task": "'"$TASK"'",
    "project": "GOTRS - OTRS replacement system",
    "constraints": [
      "Container-first development (docker/podman)",
      "TDD methodology required",
      "OTRS schema compatibility maintained", 
      "EN/DE internationalization required",
      "No OTRS licensing violations"
    ]
  }
}'
echo ""
echo "⚠️  ENFORCEMENT: No implementation may begin until workflow-orchestrator provides detailed plan"
echo "📖 Next step: Copy this context to Claude and request workflow-orchestrator consultation"