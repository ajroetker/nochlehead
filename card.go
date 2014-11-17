package main

// [0, 3]
type Player int
func (p Player) Team() int {
    return int(p) % 2
}

type Round struct {
    Sequence PlayStr `bson:"sequence" json:"sequence"`
    Trump Suit `bson:"trump" json:"trump"`
    Bidder Player `bson:"bidder" json:"bidder"`
    Bid int `bson:"bid" json:"bid"`
    Traded TradeStr `bson:"traded" json:"traded"`
}

type TradeStr string
func (t TradeStr) GetCards(n Player) [2][3]*Card {
    var cards [2][3]*Card
    for i := range cards {
        for j := range cards[i] {
            cards[i][j] = &Card{
                Value: toValue[t[j]],
                Suit: toSuit[t[j+1]],
                Owner: Player( int(n) + (i*2) ) % 4,
            }
        }
    }
    return cards
}

type PlayStr string
func (p PlayStr) GetHand(n Player, bidder Player) []*Card {
    nIndex := Player( int(bidder+n) % 4 )
    // len(PlayStr) == 96
    // In order to keep the function
    // polymorphic (taking any n | n % 8 == 0)
    // might as well do the following calculation
    hand := make([]*Card, len(p)/8)
    for handIndex := range hand {
        valueIndex := handIndex * 8 + int(nIndex)
        hand[handIndex] = &Card{
            Value: toValue[p[valueIndex]],
            Suit: toSuit[p[valueIndex+1]],
            Owner: nIndex,
        }
    }
    return hand
}

var toValue = map[uint8]Value{ 'n':1, 'j':2, 'q':3, 'k':4, 't':5, 'a':6, }
type Value int
func (v Value) String() string {
    switch v {
    case 1:
        return "n"
    case 2:
        return "j"
    case 3:
        return "q"
    case 4:
        return "k"
    case 5:
        return "t"
    case 6:
        return "a"
    default:
        return string(v)
    }
}

var toSuit = map[uint8]Suit{ 's':1, 'h':2, 'c':3, 'd':4, }
type Suit int
func (s Suit) String() string {
    // 1 := 'Spades', 2 := 'Hearts', 3 := 'Clubs', 4 := 'Diamonds'
    switch s {
    case 1:
        return "s"
    case 2:
        return "h"
    case 3:
        return "c"
    case 4:
        return "d"
    default:
        return string(s)
    }
}

type Card struct {
    Suit Suit `bson:"suit" json:"suit"`
    Value Value `bson:"value" json:"value"`
    Owner Player `bson:"owner" json:"owner"`
}

// The current round's state
type TableTop struct {
    Middle []*Card
    Current Player
    Hands [4][]*Card
    // The played cards
    Tricks [2][]*Card
    // The player who led this round
    Trump Suit
    Traded [2][3]*Card
    Melds [2]int
    Bidder Player
    Bid int
}

func (r *Round) GetHands() [4][]*Card {
    var hands [4][]*Card
    for i := range hands {
        hands[i] = r.Sequence.GetHand(Player(i), r.Bidder)
    }
    return hands
}

func TrackWinner( trump Suit, cards chan *Card) chan Player {
    winnerChan := make(chan Player)
    go func() {
        for {
            leaderCard, ok := <-cards
            if !ok {
                return
            }
            for i := 0; i < 3; i++ {
                curCard := <-cards
                if leaderCard.Suit != trump && curCard.Suit == trump {
                    leaderCard = curCard
                }
                if ! ( leaderCard.Suit == trump && curCard.Suit != trump ) {
                    if curCard.Value > leaderCard.Value {
                        leaderCard = curCard
                    }
                }
            }
            winnerChan<-leaderCard.Owner
        }
    }()
    return winnerChan
}

func (r *Round) Play() ( chan *Card, chan Player ) {
    outMove := make(chan *Card)
    outWinner := make(chan Player)
    go func() {
        moves := make(chan *Card)
        turnLength := 8
        numOfTurns := len(r.Sequence)/turnLength
        inWinner := TrackWinner(r.Trump,moves)
        first := r.Bidder
        for turn := 0; turn < numOfTurns; turn++ {
            for i := 0; i < 4; i++ {
                card := &Card{
                    Suit: toSuit[r.Sequence[(2 * (int(first) + i+1)) % 8 + 8 * turn]],
                    Value: toValue[r.Sequence[(2 * (int(first) + i)) % 8 + 8 * turn]],
                    Owner: Player( ( int(first) + i ) % 4 ),
                }
                moves<- card
                outMove<- card
            }
            first = <-inWinner
            outWinner <- first
        }
    }()
    return outMove, outWinner
}

func (tt *TableTop) Transition(move *Card, winner Player) {
    if len(tt.Middle) == 4 {
        tt.Tricks[winner.Team()] = append(tt.Tricks[winner.Team()],tt.Middle...)
        tt.Middle = make([]*Card, 0)
        tt.Current = winner
    } else {
        tt.Current = (tt.Current + 1) % 4
    }

    tt.Hands[tt.Current] = RemoveCard(tt.Hands[tt.Current], move)
    tt.Middle = append(tt.Middle, move)
}

func RemoveCard(hand []*Card, card *Card) []*Card {
    newHand := make([]*Card, 0)
    for _, member := range hand {
        if ! ( member.Suit == card.Suit && member.Value == card.Value ) {
            newHand = append(newHand, member)
        }
    }
    return newHand
}

func (r *Round) InitialConditions() *TableTop {
    tt := &TableTop{
        Bidder: r.Bidder,
        Bid: r.Bid,
        Trump: r.Trump,
        Hands: r.GetHands(),
        Current: r.Bidder,
        Middle: make([]*Card, 0),
        Traded: r.Traded.GetCards(r.Bidder),
    }
    tt.Tricks[0] = make([]*Card, 0)
    tt.Tricks[1] = make([]*Card, 0)
    tt.Melds[0] = computeMeld(tt.Trump, tt.Hands[0]) + computeMeld(tt.Trump, tt.Hands[2])
    tt.Melds[1] = computeMeld(tt.Trump, tt.Hands[1]) + computeMeld(tt.Trump, tt.Hands[3])
    return tt
}

func computeMeld(trump Suit, hand []*Card) int {
    // TODO Implement a meld function
    return classAMeld(trump, hand) + classBMeld(trump, hand) + classCMeld(trump, hand)
}

func classAMeld(trump Suit, hand []*Card) int {
    return 0
}

func classBMeld(trump Suit, hand []*Card) int {
    return 0
}

func classCMeld(trump Suit, hand []*Card) int {
    meld := 0
    queenOfSpades := 0
    jackOfDiamonds := 0
    for _, card := range hand {
        if card.Value == Value(9) && card.Suit == trump {
            meld += 10
        }
        if card.Value == Value(3) && card.Suit == Suit(1) {
            queenOfSpades++
        }
        if card.Value == Value(2) && card.Suit == Suit(4) {
            jackOfDiamonds++
        }
    }
    if queenOfSpades > 0 && jackOfDiamonds > 0 {
        if queenOfSpades == 1 {
            // Single pinochle
            meld += 40
        } else {
            if jackOfDiamonds == 1 {
            } else {
                // Double pinochle
                meld += 300
            }
        }
    }
    return meld
}

type CompiledGame struct {
    Rounds []*Round `bson:"rounds" json:"rounds"`
    Names  [4]string `bson:"names" json:"names"`
}

func (c *CompiledGame) Name( p Player ) string {
    return c.Names[p]
}

func (c *CompiledGame) RoundChan() chan *Round {
    out := make(chan *Round)
    go func() {
        for _, round := range c.Rounds {
            out <- round
        }
    }()
    return out
}
