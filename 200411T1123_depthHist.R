p <- plot_ly()


chr3 <- list(line = list(shape = "spline"),mode = "lines+markers",name = "chr3",type = "scatter",
y = c(0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 1.00, ),
x = c(1, 2, 3, 4, 5, 6, 7, 8),
hovertemplate = "Depth: %{x}")
p <- add_trace(p, line=chr3$line, mode=chr3$mode, name=chr3$name, type=chr3$type, x=chr3$x, y=chr3$y, text=chr3$text, hovertemplate=chr3$hovertemplate)