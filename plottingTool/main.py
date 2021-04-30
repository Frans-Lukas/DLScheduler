import plotly.graph_objects as go

if __name__ == '__main__':
    print("fafaf")
    fig = go.Figure(data=go.scatter(y=[2, 3, 1], x=[2, 4, 1]))
    fig.write_html('first_figure.html', auto_open=True)
