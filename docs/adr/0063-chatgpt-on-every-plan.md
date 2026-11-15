# 63. ChatGPT publishes on every plan, and the first steps no longer ask for its plan

- Status: Accepted
- Date: 2026-11-15

## Context

ChatGPT is the second AI client of Cahut, through a custom MCP server that a person adds on the web ([ADR 29](0029-chatgpt-second-ai-client.md)). In October 2026, the help of OpenAI kept that server to the Business, Enterprise and Edu plans, and gave Pro a server for reading only. The first steps of the console therefore ask for the plan of ChatGPT first, and with Free, Go, Plus or Pro point to Claude and Mistral Vibe instead ([ADR 61](0061-a-guide-to-the-first-ai.md)). The documentation, the website and the README say the same.

Later in October 2026, the team of OpenAI that builds plugins announced that a custom MCP server no longer needs developer mode, and that every person adds one, from Plugins, Add. An account on the Free plan adds a Cahut there and publishes to it. Where the item is missing, turning on developer mode under Settings, Security and login shows it. The help article of OpenAI still lists the plans of organizations only.

Most people who use ChatGPT pay for no plan, or for one of their own. The guide sent them away from an AI that works.

Gemini, in its web app, adds a custom MCP server as a custom app, under Settings & help, Connected Apps. Google keeps it to people in the United States, from 18, with a personal Google Account. Nothing in Cahut names it yet.

## Decision

- The first steps no longer ask for the plan of ChatGPT: its page gives its steps at once. The choice of ChatGPT says that it is on the web, free plan included.
- The steps of ChatGPT, in the guide and in **AIs and access**, say that every plan adds a custom MCP server, that developer mode shows the item where it is missing, and that in a workspace its administrators choose who adds one.
- The documentation, the website and the README list ChatGPT among the clients that publish on their free plans, and Gemini with its conditions. They name the help of each vendor and its date, and that help has the last word.
- The guide does not name Gemini: **Another AI** gives the address of the Cahut, which Gemini takes as a custom app.
- The answer for people with no AI yet still points to Claude and Mistral Vibe, which a phone connects, where those people often are.

## Consequences

- A person on ChatGPT Free, Go, Plus or Pro connects it from the guide, as on the other plans.
- Should OpenAI keep custom MCP servers to some plans again, the words of the steps say so, without a question that turns people away.
- Gemini joins the guide, with steps of its own, once Google opens custom apps beyond the United States and a Cahut has been tried with it.

## Alternatives considered

- **Keeping the question, with every plan as an answer.** A question that changes nothing slows the first step.
- **Following the help article of OpenAI, which still lists Business, Enterprise and Edu.** It turns away most people who use ChatGPT, while their accounts publish.
- **Giving Gemini its own page in the guide now.** It serves people in the United States only, and has not been tried with a Cahut yet.
