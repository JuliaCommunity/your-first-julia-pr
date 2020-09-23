using Luxor, ColorTypes

const JULIA_BLUE    = RGB(0.251, 0.388, 0.847);
const JULIA_GREEN   = RGB(0.22 , 0.596, 0.149);
const JULIA_PURPLE  = RGB(0.584, 0.345, 0.698);
const JULIA_RED     = RGB(0.796, 0.235, 0.2  );

const JULIA_COLOURS = [JULIA_BLUE, JULIA_GREEN, JULIA_PURPLE, JULIA_RED];

const size_x = 2560
const size_y = 1440

const r_min =   2
const r_max = 100

# const n_circles = 50000
const attempts  = 500

mutable struct C
    x::Int
    y::Int
    r::Int
end

function collides(c₁::C, circles)
    for c₂ in circles
        a = c₁.r + c₂.r
        x = c₁.x - c₂.x
        y = c₁.y - c₂.y

        a >= sqrt(x^2 + y^2) && return true
    end

    (c₁.x + c₁.r >= size_x || c₁.x - c₁.r <= 0) && return true
    (c₁.y + c₁.r >= size_y || c₁.y - c₁.r <= 0) && return true

    return false
end

function main(n_circles)
    Drawing(size_x, size_y, "background.svg")
    circles = C[]

    for i in 1:n_circles
        local c
        safe = false

        for j in 1:attempts
            c = C(floor(rand() * size_x), floor(rand() * size_y), r_min)

            if !collides(c, circles)
                safe = true
                break
            end
        end

        safe || continue

        for r in r_min:r_max
            c.r = r
            if collides(c, circles)
                c.r -= 1
                break
            end
        end

        push!(circles, c)

        sethue(rand(JULIA_COLOURS))
        circle(c.x, c.y, c.r, :fill)
    end

    finish()
    preview()
end

main(3) # compile
main(50000) # run