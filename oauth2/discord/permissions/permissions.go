package permissions

import "encoding/json"

type Perm struct {
	CreateInstantInvite bool
	KickMembers         bool
	BanMembers          bool
	ManageRoles         bool
	ManageChannels      bool
	ManageGuild         bool

	// Text
	ReadMessages       bool
	SendMessages       bool
	SendTssMessages    bool
	ManageMessages     bool
	EmbedLinks         bool
	AttachFiles        bool
	ReadMessageHistory bool
	MentionEveryone    bool

	// Voice
	Connect       bool
	Speak         bool
	MuteMembers   bool
	DeafenMembers bool
	MoveMembers   bool
	UseVad        bool
}

func (p Perm) ToInt() int {
	i := 0
	if p.CreateInstantInvite {
		i |= 1 << 0
	}
	if p.KickMembers {
		i |= 1 << 1
	} // not in channel overwrites
	if p.BanMembers {
		i |= 1 << 2
	} // not in channel overwrites
	if p.ManageRoles {
		i |= 1 << 3
	}
	if p.ManageChannels {
		i |= 1 << 4
	}
	if p.ManageGuild {
		i |= 1 << 5
	} // not in channel overwrites (understandably)

	// Text
	if p.ReadMessages {
		i |= 1 << 10
	}
	if p.SendMessages {
		i |= 1 << 11
	}
	if p.SendTssMessages {
		i |= 1 << 12
	}
	if p.ManageMessages {
		i |= 1 << 13
	}
	if p.EmbedLinks {
		i |= 1 << 14
	}
	if p.AttachFiles {
		i |= 1 << 15
	}
	if p.ReadMessageHistory {
		i |= 1 << 16
	}
	if p.MentionEveryone {
		i |= 1 << 17
	}

	// Voice
	if p.Connect {
		i |= 1 << 20
	}
	if p.Speak {
		i |= 1 << 21
	}
	if p.MuteMembers {
		i |= 1 << 22
	}
	if p.DeafenMembers {
		i |= 1 << 23
	}
	if p.MoveMembers {
		i |= 1 << 24
	}
	if p.UseVad {
		i |= 1 << 25
	}
	return i
}
func (p *Perm) FromInt(mask int) {
	if (1 << 0 & mask) != 0 {
		p.CreateInstantInvite = true
	}
	if (1 << 1 & mask) != 0 {
		p.KickMembers = true
	}
	if (1 << 2 & mask) != 0 {
		p.BanMembers = true
	}
	if (1 << 3 & mask) != 0 {
		p.ManageRoles = true
	}
	if (1 << 4 & mask) != 0 {
		p.ManageChannels = true
	}
	if (1 << 5 & mask) != 0 {
		p.ManageGuild = true
	}

	// Text
	if (1 << 10 & mask) != 0 {
		p.ReadMessages = true
	}
	if (1 << 11 & mask) != 0 {
		p.SendMessages = true
	}
	if (1 << 12 & mask) != 0 {
		p.SendTssMessages = true
	}
	if (1 << 13 & mask) != 0 {
		p.ManageMessages = true
	}
	if (1 << 14 & mask) != 0 {
		p.EmbedLinks = true
	}
	if (1 << 15 & mask) != 0 {
		p.AttachFiles = true
	}
	if (1 << 16 & mask) != 0 {
		p.ReadMessageHistory = true
	}
	if (1 << 17 & mask) != 0 {
		p.MentionEveryone = true
	}

	// Voice
	if (1 << 20 & mask) != 0 {
		p.Connect = true
	}
	if (1 << 21 & mask) != 0 {
		p.Speak = true
	}
	if (1 << 22 & mask) != 0 {
		p.MuteMembers = true
	}
	if (1 << 23 & mask) != 0 {
		p.DeafenMembers = true
	}
	if (1 << 24 & mask) != 0 {
		p.MoveMembers = true
	}
	if (1 << 25 & mask) != 0 {
		p.UseVad = true
	}
}
func FromInt(mask int) (p Perm) {
	p.FromInt(mask)
	return
}

func (p Perm) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToInt())
}
func (p *Perm) UnmarshalJSON(raw []byte) error {
	var i int
	err := json.Unmarshal(raw, &i)
	if err != nil {
		return err
	}
	p.FromInt(i)
	//fmt.Printf("int: %v\nperm:\n%+v\n\n", i, p)
	return nil
}
