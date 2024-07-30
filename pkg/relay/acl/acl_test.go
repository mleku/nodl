package acl

var testRelaySec = "f16dca5c36931305a4ac30d31b77962af96ea6b7240736da11af318fb7e11317"

// func TestT(t *testing.T) {
// 	var signers []*p256k.Signer
// 	var err error
// 	for _ = range 10 {
// 		var skb B
// 		if skb, _,_,_, err = p256k.GenSecBytes(); chk.E(err) {
// 			t.Fatal(err)
// 		}
// 		signer := &p256k.Signer{}
// 		if err = signer.InitSec(skb); chk.E(err) {
// 			t.Fatal(err)
// 		}
// 		signers = append(signers, signer)
// 	}
// 	log.I.S(signers)
// 	aclT := &T{}
// 	for i := range signers {
// 		role := (i % (len(RoleStrings) - 1)) + 1
// 		en := &Entry{
// 			Role:         Role(role),
// 			Pubkey:       signers[i].Pub(),
// 			Created:      timestamp.FromUnix(timestamp.Now().I64() - 1),
// 			LastModified: timestamp.Now(),
// 			Expires:      timestamp.FromUnix(timestamp.Now().I64() + 100000),
// 		}
// 		if err = aclT.AddEntry(en); err != nil {
// 			t.Fatal(err)
// 		}
// 		ev := en.ToEvent()
// 		signer := &p256k.Signer{}
// 		if err = signer.InitPub(en.Pubkey); chk.E(err) {
// 			t.Fatal(err)
// 		}
// 		if err = ev.Sign(signer); err != nil {
// 			t.Fatal(err)
// 		}
// 		var e *Entry
// 		if e, err = aclT.FromEvent(ev); err != nil {
// 			t.Fatal(err)
// 		}
// 		_ = e
// 	}
// 	frand.Shuffle(len(signers), func(i, j int) {
// 		signers[i], signers[j] = signers[j], signers[i]
// 	})
// 	for i := range signers {
// 		if err = aclT.DeleteEntry(signers[i].Pub()); err != nil {
// 			t.Fatal(err)
// 		}
// 	}
// }
