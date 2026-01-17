package mc_test

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/camry/mc"
)

type UserProp struct {
}

type GuildMemberInfo struct {
}

func TestGetStructName(t *testing.T) {
    assert.Equal(t, "UserProp", mc.GetStructName(&UserProp{}))
    assert.Equal(t, "UserProp", mc.GetStructName(UserProp{}))
    assert.Equal(t, "GuildMemberInfo", mc.GetStructName(&GuildMemberInfo{}))
    assert.Equal(t, "GuildMemberInfo", mc.GetStructName(GuildMemberInfo{}))
}
