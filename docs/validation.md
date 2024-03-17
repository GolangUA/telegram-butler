# Validation

## The requirements:

- Filter out bots.
- Filter out unloyal to the Ukraine users. (russians, etc...)

## Proposal #1

Generally the Message struct is huge, but it doesn't have a lot useful information available:

```go
type Message struct {
	// MessageID - Unique message identifier inside this chat
	MessageID int `json:"message_id"`

	// MessageThreadID - Optional. Unique identifier of a message thread to which the message belongs; for
	// supergroups only
	MessageThreadID int `json:"message_thread_id,omitempty"`

	// From - Optional. Sender of the message; empty for messages sent to channels. For backward compatibility,
	// the field contains a fake sender user in non-channel chats, if the message was sent on behalf of a chat.
	From *User `json:"from,omitempty"`

	// SenderChat - Optional. Sender of the message, sent on behalf of a chat. For example, the channel itself
	// for channel posts, the supergroup itself for messages from anonymous group administrators, the linked channel
	// for messages automatically forwarded to the discussion group. For backward compatibility, the field from
	// contains a fake sender user in non-channel chats, if the message was sent on behalf of a chat.
	SenderChat *Chat `json:"sender_chat,omitempty"`

	// SenderBoostCount - Optional. If the sender of the message boosted the chat, the number of boosts added by
	// the user
	SenderBoostCount int `json:"sender_boost_count,omitempty"`

	// Date - Date the message was sent in Unix time. It is always a positive number, representing a valid date.
	Date int64 `json:"date"`

	// Chat - Chat the message belongs to
	Chat Chat `json:"chat"`

	// ForwardOrigin - Optional. Information about the original message for forwarded messages
	ForwardOrigin MessageOrigin `json:"forward_origin,omitempty"`

    // And a lot more, but it's useless in our case
    // ...
}
```

Using this approach we'll get the following flow:

GolangUA group is totally private, no way to join with some link, etc. -> User send a command DIRECTLY to our bot. (i.E /join) -> Bot perform checks on a nickname (russian validation: checking on forbidden words/emojis etc.) -> Bot generates invite link to the GolangUA with limit on join request -> Bot sends a following or a similar message: "Clicking this button you agree that Russia is aggressor, Crimea and Donbass is Ukraine and Putin is khyilo" and attaches interactive button with invite link -> User click this button and successfully pases all the checks and going directly to the group (we performed two checks in one, checking if this a bot by asking for a button click, and by clicking a button we assume that user agree with over terms and supports Ukraine)

Pros:

- All checks will be private to user. We're not going to spam somewhere in the GolangUA group.
- We can go crazy with all this checks directly in the bot. We can add more complicated checks after some time

Cons:

- The flow is quite complicated comparing to the second option

## Proposal #2

Information available:

```go
// ChatJoinRequest - Represents a join request sent to a chat.
type ChatJoinRequest struct {
	// Chat - Chat to which the request was sent
	Chat Chat `json:"chat"`

	// From - User that sent the join request
	From User `json:"from"`

	// UserChatID - Identifier of a private chat with the user who sent the join request. This number may have
	// more than 32 significant bits and some programming languages may have difficulty/silent defects in
	// interpreting it. But it has at most 52 significant bits, so a 64-bit integer or double-precision float type
	// are safe for storing this identifier. The bot can use this identifier for 5 minutes to send messages until
	// the join request is processed, assuming no other administrator contacted the user.
	UserChatID int64 `json:"user_chat_id"`

	// Date - Date the request was sent in Unix time
	Date int64 `json:"date"`

	// Bio - Optional. Bio of the user.
	Bio string `json:"bio,omitempty"`

	// InviteLink - Optional. Chat invite link that was used by the user to send the join request
	InviteLink *ChatInviteLink `json:"invite_link,omitempty"`
}

```

Using this approach we'll get the following flow:

User send a join request -> Bot perform the same checks on a nickname/bio as in previous proposal (russian validation) -> If it's OK accept it -> Perform additional validation: ask to calculate a simple equation or give a response to the dummy question (bot validation) directly in GolangUA group

Pros:

- Less time to implement.
- We have access to the user bio without any addition requests. Therefore we can perform russian validation based on bio too.

Cons:

- Bot validation (dummy questions etc) will be visible to other users
- We need to define a timeout for bot validation. I.E: if user hadn't respond during 15 min after bot accepted his request he'll be kicked. Of course, this type of timeout will have a lot of false positives.
