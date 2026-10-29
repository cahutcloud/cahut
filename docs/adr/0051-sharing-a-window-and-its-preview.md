# 51. People share a Window by its permanent address, which shows the robots of previews a card of Cahut

- Status: Accepted
- Date: 2026-10-29

## Context

People share a Window by pasting its address in a message, a post or an e-mail. Slack, Teams, Discord, iMessage, WhatsApp, X, LinkedIn and Facebook then show a preview of the link: a robot fetches the address, reads the tags of Open Graph (`og:title`, `og:description`, `og:image`, `og:url`) and of X (`twitter:card`) in the page it gets, and draws a title, a description and a picture under the link. The robot has no session and no Key. iMessage and WhatsApp fetch from the phone of the person who sends the link, the others from their own servers.

Today a robot that follows the address of a Window gets:

- for a public Window, the `index.html` of its author, with the tags that its AI wrote: often none but a `<title>`, and no picture;
- for any other Window, the page of what is not found, `404 Not Found`, or the page of its Key, `401 Unauthorized`, which shows the title of the Window ([ADR 19](0019-window-versions-access-and-declaration.md)). Most robots draw no preview of an error, and the link shows bare.

[ADR 50](0050-window-ids-permanent-addresses-and-names.md) gave each Window a permanent address, `https://<domain>/-/w/<id>`, which redirects a reader who may open the Window to its address of the moment, and gives the others the page of what is not found or of the Key. It left the preview of a link to this record, with `og:url` naming the permanent address.

Olivier decided on 29 October 2026 that:

- people share a Window, not the Cahut;
- a button Share opens a menu on the page of a Window in the console and on the page of a Room; there is no button in the Window, which belongs to its author;
- the entries of the menu are plain links: `mailto:` and the addresses of sharing of X, LinkedIn and Facebook, with no script or tracker of theirs;
- `og:title` and `og:description` are the title and the description of the Window, and `og:url`, like the address shared, is its permanent address;
- `og:image` is a card in the brand of Cahut, with its logo, the title of the Window and the name of the Cahut, which the server draws without a browser.

This record settles the rest.

## Decision

### What a preview says

- A preview tells no more than its address tells a stranger, who has no session and no Key:

  | Access of the Window | `og:title` | `og:description` | Card |
  | --- | --- | --- | --- |
  | `public` | Its title | Its description, when it has one | With its title |
  | `key` | Its title, which the page of its Key already shows | Opens with its Key. | With its title |
  | `room`, `internal` | A Window at `<domain>` | Sign in to open it. | Without a title |

- The access that counts is the one of the moment. An access that ends falls back to `room`, and the preview with it, for the next robot that comes. The previews that platforms already drew stay as they are in their messages: Cahut cannot reach them.
- A Window in the trash or purged, and an id that names no Window, give the page of what is not found, as today, and no preview.

### Where the tags live

- **At the permanent address, Cahut answers the robots of previews with a page of its own**: `200 OK`, a page of Cahut in its brand, whose head holds the tags, and whose body gives the same title and description, with a link to the permanent address.
  - `og:type` `website`, `og:site_name` the domain of the Cahut, `og:title`, `og:description`, `og:url` the permanent address, `og:image` the card, with `og:image:width` 1200, `og:image:height` 630 and `og:image:alt`, and `twitter:card` `summary_large_image`, which X needs to show a large picture.
  - The robots are known by their `User-Agent`: a list in the code of the words that the robots of previews put in it, such as `Slackbot`, `Twitterbot`, `facebookexternalhit` (which iMessage sends too), `LinkedInBot`, `Discordbot`, `TelegramBot`, `WhatsApp`, `SkypeUriPreview` and `MicrosoftPreview` (Teams), `Mastodon` and `Cardyb` (Bluesky). Search engines are not on it: they follow the redirect, as browsers do.
  - The page of a robot opens nothing: no session, no pass, no frame, and no Window opens for having been previewed.
- **Everyone else gets the permanent address of ADR 50**, unchanged: the redirect for a reader who may open the Window, and the page of what is not found or of the Key for the others.
- **The address of the moment does not change.** It serves the files of a public Window as its AI wrote them, its own tags included, and the pages of ADR 19 for the others. Cahut writes nothing into the files of a Window.
- A robot that is not on the list, and a link to the address of the moment, get what a person without a session gets: the tags of the author for a public Window, an error for the others. The list grows as robots are found.

### The card

- **Its address is `/-/card/<id>.png`.** `og:image` adds `?v=` and a short hash of what the card shows, so that a platform that keeps pictures by address takes the new card after a new title. The card itself always shows the Window as it is.
- **It shows what the preview shows**: the title for a public Window or a Window with a Key, none for the others. Anyone who has the id can fetch it, as the robots do.
- **It is drawn on demand and stored nowhere.** The server keeps the last few dozen cards it drew in memory, keyed by what they show, and serves them with an `ETag` and `Cache-Control: public, max-age=3600`. A card costs about 30 milliseconds to draw, and a new title or access a new card.
- **It follows the picture of sharing of the brand guidelines** ([`og-image.png`](../charte-graphique/applications/og-image.png)), 1200 by 630 pixels, in PNG: Paper, the horizontal logo at the top left, the title of the Window in Plus Jakarta Sans ExtraBold in Ink where the signature stands, the domain of the Cahut in IBM Plex Mono in Sky 700 at the bottom left, and the open house on the right. A card without a title keeps the rest.
- **A long title shrinks, then stops**: 64 pixels on three lines at most, then 56, 48 and 40 pixels on four or five lines, and at the last size an ellipsis ends the last line. Lines break between words, and inside a word too long for a line.
- **Scripts.** Plus Jakarta Sans draws the Latin script, Vietnamese included, but not Greek, Cyrillic, Hebrew, Arabic, nor the scripts of India or of East Asia. The title is first normalized (NFKC, so that `2ᵉ` reads `2e`), and the symbols that the font lacks, such as emoji, are left out. When the font lacks a letter or a digit of the title, the card shows no title: the preview still gives it in `og:title`, which the platforms write in their own fonts, shaped and in its direction.
- **It is drawn in Go, without cgo**, with `golang.org/x/image`, a module of the Go project in pure Go: `font/opentype` reads the fonts and draws the text, with the kerning of the fonts, and `draw` scales the pictures. The binary embeds, as they are in the brand guidelines, the TrueType fonts `PlusJakartaSans-ExtraBold.ttf` and `IBMPlexMono-Regular.ttf`, and the PNG renders of the horizontal logo and of the open house, about 400 KB in all. The logo is never redrawn.

### The menu Share

- **Where.** A button Share, beside Open, at the head of the page of a Window in the console; and a button Share on each card of a Window in the console, beside the one that opens it, on the page of a Room and on the overview, which show the same cards. None on the pages that readers see, and none in the Window.
- **What it shares.** The permanent address, with the title of the Window, whatever its access.
- **Its entries:**

  | Entry | What it does |
  | --- | --- |
  | Copy the link | Copies the permanent address |
  | E-mail | `mailto:?subject=<title>&body=<description>%0D%0A%0D%0A<permanent address>`: the mail app of the person opens, with no words of Cahut |
  | X | `https://x.com/intent/tweet?text=<title>&url=<permanent address>` |
  | LinkedIn | `https://www.linkedin.com/sharing/share-offsite/?url=<permanent address>`, its only parameter |
  | Facebook | `https://www.facebook.com/sharer/sharer.php?u=<permanent address>` |
  | Share… | The share sheet of the system, where the browser has one (`navigator.share`): phones, Safari, Edge |

  Values are encoded with `encodeURIComponent`. X, LinkedIn and Facebook open in a new tab, with `rel="noopener noreferrer"`: the console loads nothing of theirs, and they learn of the Window when the person posts it.
- **Who opens the link.** The menu says it from the access: anyone, anyone with its Key, the members of its Room, or the people of the Cahut. X, LinkedIn and Facebook show for a public Window and a Window with a Key only: a post about a Window open to its Room or its Cahut would reach people who cannot open it.
- **Nothing goes to the Logbook.** The menu changes nothing in the Cahut, and Cahut does not know whether the person posted.

### Later

- The secret and provisional links of sharing, and the Windows that open once, have their own addresses and their own records. Those records say what their previews show, from the same page and card; a robot that previews a link that opens once must not use it up, and the page of a robot opens nothing.
- A button Share on the page of a Room that readers see, and the sharing of a Room, once Rooms have a permanent address.

## Consequences

- A link to a Window shows a preview in the brand of Cahut wherever people paste it, with its title and its description when the Window is public, and its title when it has a Key.
- The title of a Window open to its Room or its Cahut stays out of the previews, of the servers of the platforms that draw them, and of whoever gets hold of the link.
- The robots and the people who follow the same permanent address get different answers. Search engines get what people get.
- The list of robots needs care: a robot that is new, or that changes its `User-Agent`, gets what a person without a session gets.
- A preview cannot be taken back: a title shared while its Window was public stays in the messages that show it.
- A public Window has two previews: the one of Cahut at its permanent address, and the one of its author at its address of the moment. The menu gives the first.
- `golang.org/x/image` becomes a dependency, and `golang.org/x/text`, already in the build, a direct one. The binary grows by the fonts and the pictures of the card.
- The brand guidelines show the card among their applications.
- This record extends [ADR 50](0050-window-ids-permanent-addresses-and-names.md) (the permanent address answers the robots of previews with a page of Cahut) and [ADR 27](0027-previews-of-windows-in-the-console.md) (the cards of Windows in the console have a button Share).

## Alternatives considered

- **The tags written into the HTML of the author**, as Cahut serves it. Any robot would find them, at either address, but only for public Windows; Cahut would change the files of a Window, and would have to choose between the tags of its author and its own. With `og:url` at the permanent address, Facebook would go round in a circle, from the permanent address to the Window and back.
- **The tags on the pages that a visitor without access gets**, those of what is not found and of the Key. They answer `404` and `401`, which most robots do not preview; answering `200` would tell whoever guesses the name of a Room and of a Window that it exists, which ADR 19 rules out. And a public Window has no page of Cahut to carry them.
- **The robots known by `Sec-Fetch-Dest`**, which browsers send and robots do not, rather than by a list. No list to keep, but an old browser or the web view of an app would get the page of a robot rather than the Window.
- **The title in the preview of every Window**, whatever its access. Simpler, but whoever gets hold of the link, and every platform that draws its preview, reads the title of a Window open to its Room only.
- **No preview for the Windows that are not public.** Nothing to say, but a link shared within a team, in Slack or Teams, shows bare.
- **The card drawn at publication and stored** with the Window. It would have to follow each title and each access, would count in the disk of the Cahut, and would be drawn for Windows that nobody shares.
- **The card drawn by a headless browser** from a template of the brand guidelines: against the single static binary, as [ADR 27](0027-previews-of-windows-in-the-console.md) found for previews.
- **Fonts that cover more scripts**, such as Noto, with a shaping engine such as `go-text/typesetting`: megabytes more in the binary, the Chinese and Japanese fonts alone weigh over 15 MB, for a picture whose title the preview already gives in text.
- **A button Share on the page of a Room that readers see.** Its readers share the address they read, and its pages run as little script as they can.
