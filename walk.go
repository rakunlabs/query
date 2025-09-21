package query

type WalkType int

const (
	WalkCurrent WalkType = iota
	WalkStart
	WalkEnd
)

type stackHolder struct {
	Current  []Expression
	Position int
}

type Token struct {
	Expression Expression
	Type       WalkType
}

func newToken(expr Expression, typ WalkType) Token {
	return Token{
		Expression: expr,
		Type:       typ,
	}
}

// Walk traverses the query tree and calls the provided function for each expression.
func (q *Query) Walk(fn func(Token) error) error {
	var stack []*stackHolder

	stack = append(stack, &stackHolder{
		Current:  q.Where,
		Position: 0,
	})

	for {
		currentStack := stack[len(stack)-1]

		// Check if we need to pop the stack, end of current stack
		if currentStack.Position >= len(currentStack.Current) {
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				if err := fn(newToken(&ExpressionLogic{Operator: OperatorAnd}, WalkEnd)); err != nil {
					return err
				}

				break
			}
			currentStack = stack[len(stack)-1]
			if err := fn(newToken(currentStack.Current[currentStack.Position], WalkEnd)); err != nil {
				return err
			}
			currentStack.Position++

			continue
		}

		expr := currentStack.Current[currentStack.Position]
		if e, ok := expr.(*ExpressionLogic); ok {
			stack = append(stack, &stackHolder{
				Current:  e.List,
				Position: 0,
			})
			if err := fn(newToken(e, WalkStart)); err != nil {
				return err
			}

			continue
		}

		if err := fn(newToken(expr, WalkCurrent)); err != nil {
			return err
		}

		currentStack.Position++
	}

	return nil
}
